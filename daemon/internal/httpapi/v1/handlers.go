package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmriccardo/backer/deamon/internal/utils"
)

func HandleListJobs(ctx *gin.Context) {

}

func HandleJobCreateRequest(ctx *gin.Context) {
	// Binds the request body to the request structure and applies
	// defaults where necessary
	var req CreateJobRequest
	if err := utils.MustBindWithJSON(&req, ctx.Request); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

// RegisterHandlers registers v1 handlers
func RegisterHandlers(rg *gin.RouterGroup) {
	rg.Group("/v1").Group("/jobs").
		GET("/", HandleListJobs).
		POST("/create", HandleJobCreateRequest)
}
