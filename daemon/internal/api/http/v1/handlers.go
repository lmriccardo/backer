package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmriccardo/backer/deamon/internal/app/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
)

func HandleListJobs(ctx *gin.Context, srv *service.Service) {

}

func HandleJobCreateRequest(ctx *gin.Context, srv *service.Service) {
	// Binds the request body to the request structure and applies
	// defaults where necessary
	var req CreateJobRequest
	if err := utils.MustBindWithJSON(&req, ctx.Request); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validates the request payload using internal gin validator
	if ok := ValidateRequest(req, ctx); !ok {
		return
	}

	if err := srv.CreateJob(ctx.Request.Context(), &req, nil); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

// RegisterHandlers registers v1 handlers
func RegisterHandlers(rg *gin.RouterGroup, srv *service.Service) {
	_ = RegisterValidators() // Registers all validators

	rg.Group("/v1").Group("/jobs").
		GET("/", func(ctx *gin.Context) { HandleListJobs(ctx, srv) }).
		POST("/create", func(ctx *gin.Context) { HandleJobCreateRequest(ctx, srv) })
}
