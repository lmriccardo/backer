package docs

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RunDocsServer(addr string, errChan chan error) *http.Server {
	r := gin.New()
	r.Use(gin.Recovery())

	// Swagger UI (served on TCP)
	r.GET("/backerd/api/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Printf("API Documentation at http://%s/backerd/api/index.html", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
			return
		}

		errChan <- nil
	}()

	return srv
}
