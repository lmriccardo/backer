package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmriccardo/backer/deamon/internal/core"
	"github.com/lmriccardo/backer/deamon/internal/version"
)

func healthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"version": version.Get(),
	})
}

func _version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, version.Get())
}

func NewEngine(s *core.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Add methods that do not depends on the current API version
	api := r.Group("/api")
	api.GET("/healthz", healthz)
	api.GET("/version", _version)

	return r
}
