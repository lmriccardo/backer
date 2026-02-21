package v1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmriccardo/backer/deamon/internal/app/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
)

type HandlerFunc = func(ctx *gin.Context, srv *service.Service)

func WrapHandler(srv *service.Service, fn HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) { fn(ctx, srv) }
}

func HandleListJobs(ctx *gin.Context, srv *service.Service) {

}

func HandleJobCreateRequest(ctx *gin.Context, srv *service.Service) {
	// Binds the request body to the request structure and applies
	// defaults where necessary
	log.Println("[NEW REQUEST] Received new job creation request")

	var req CreateJobRequest
	if err := utils.MustBindWithJSON(&req, ctx.Request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validates the request payload using internal gin validator
	if ok := ValidateRequest(req, ctx); !ok {
		return
	}

	// Create the job and
	if err := srv.CreateJob(ctx.Request.Context(), &req, nil); err != nil {
		// If the error returned by the function regards the job
		// with that name already existing, then we need to return 409
		status_code := http.StatusBadRequest
		if _, ok := err.(*service.DuplicateJobNameError); ok {
			status_code = http.StatusConflict
		}
		ctx.JSON(status_code, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Created job %v", req.Name),
	})
}

// RegisterHandlers registers v1 handlers
func RegisterHandlers(rg *gin.RouterGroup, srv *service.Service) {
	_ = RegisterValidators() // Registers all validators

	rg.Group("/v1").Group("/jobs").
		GET("/", WrapHandler(srv, HandleListJobs)).
		POST("/create", WrapHandler(srv, HandleJobCreateRequest))
}
