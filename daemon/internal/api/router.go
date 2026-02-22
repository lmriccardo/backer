package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "github.com/lmriccardo/backer/deamon/internal/api/v1"
	"github.com/lmriccardo/backer/deamon/internal/app/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/version"
)

// @name HealthzResponse
type HealthzResponse struct {
	Ok      bool         `json:"ok"`
	Version version.Info `json:"version"`
}

type ApiRegisterer func(*gin.RouterGroup, *service.Service)

var API_VERSION_HANDLER = map[int]ApiRegisterer{
	1: v1.RegisterHandlers,
}

// @Summary Health Check
// @Description Returns service health status and version
// @Tags system
// @Produce json
// @Success 200 {object} HealthzResponse
// @Router /healthz [get]
func healthz(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthzResponse{
		Ok:      true,
		Version: version.Get(),
	})
}

// @Summary Get Service Version
// @Description Returns current running service and api version
// @Tags system
// @Produce json
// @Success 200 {object} version.Info
// @Router /version [get]
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
	API_VERSION_HANDLER[version.API_VERSION](api, s)

	return r
}
