// @title Backer API
// @version 1.0
// @description Backup automation daemon API
// @BasePath /api

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmriccardo/backer/deamon/docs"
	"github.com/lmriccardo/backer/deamon/internal/api"
	"github.com/lmriccardo/backer/deamon/internal/app/runners"
	"github.com/lmriccardo/backer/deamon/internal/app/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/constants"
	"github.com/lmriccardo/backer/deamon/internal/platform/version"
	"github.com/lmriccardo/backer/deamon/internal/transport"
)

func main() {
	fmt.Println((version.Get()).String())

	// Create the interrupt context
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	// First create the route engine for the REST API
	service, err := service.NewService(ctx, constants.NOF_WORKERS_DEFAULT)
	if err != nil {
		log.Fatalf("when creating the service: %v", err)
	}

	engine := api.NewEngine(service)

	// Then create the server listener based on the current OS
	t, err := transport.NewTransport()
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{Handler: engine}

	// Start the server in a separated go routing and catch
	// errors on a separated channel
	serverErr := make(chan error, 1)

	// Runs the server documentation
	docsrv := docs.RunDocsServer("127.0.0.1:8081", serverErr)

	// Create all the runners with the running engine
	rChannel := service.IRegistry.GetTaskChannel()
	rEngine := runners.NewRunningEngine(ctx, constants.NOF_WORKERS_DEFAULT, rChannel)
	rEngine.Run()

	go func() {
		log.Printf("Listening on %s\n", t.Uri())
		err := srv.Serve(t.Listener)

		// Serve returns ErrServerClosed on normal shutdown
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	// Wait either for the Interrupt signal or any
	// server error. For server errors raise a fatal
	select {
	case <-ctx.Done():
		log.Println("Received shutdown signal")
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}

	// Graceful shutdown
	sdwnCtx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)

	defer cancel()

	_ = srv.Shutdown(sdwnCtx)    // stops HTTP, drains requests
	_ = docsrv.Shutdown(sdwnCtx) // Stop the HTTP Documentation server
	t.Close()                    // cleanup (listener close + unlink etc.)
	service.Close()              // Cleanup the service
	rEngine.Stop()               // Stop the runner engine

	// Wait for Serve goroutine to finish before exiting main
	if err := <-serverErr; err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Shutdown complete")
}
