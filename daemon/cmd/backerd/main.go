package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmriccardo/backer/deamon/internal/httpapi"
	"github.com/lmriccardo/backer/deamon/internal/transport"
)

func main() {
	// First create the route engine for the REST API
	engine := httpapi.NewEngine()

	// Then create the server listener based on the current OS
	t, err := transport.NewTransport()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s\n", t.Uri())
	srv := &http.Server{Handler: engine}

	// Start the server in a separated go routing and catch
	// errors on a separated channel
	serverErr := make(chan error, 1)
	go func() {
		err := srv.Serve(t.Listener)
		// Serve returns ErrServerClosed on normal shutdown
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

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

	_ = srv.Shutdown(sdwnCtx) // stops HTTP, drains requests
	t.Close()                 // your cleanup (listener close + unlink etc.)

	// Wait for Serve goroutine to finish before exiting main
	if err := <-serverErr; err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Shutdown complete")
}
