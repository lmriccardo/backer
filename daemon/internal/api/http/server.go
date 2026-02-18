package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmriccardo/backer/deamon/internal/api/http/v1"
	"github.com/lmriccardo/backer/deamon/internal/app/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/constants"
	"github.com/lmriccardo/backer/deamon/internal/platform/version"
)

type ApiRegisterer func(*gin.RouterGroup, *service.Service)

var API_VERSION_HANDLER = map[int]ApiRegisterer{
	1: v1.RegisterHandlers,
}

func healthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"version": version.Get(),
	})
}

func _version(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, version.Get())
}

func NewEngine(s *service.Service) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// Add methods that do not depends on the current API version
	api := r.Group("/api")
	api.GET("/healthz", healthz)
	api.GET("/version", _version)

	// Registers the handlers given the API version
	API_VERSION_HANDLER[constants.API_VERSION](api, s)

	return r
}
