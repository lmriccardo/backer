package v1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lmriccardo/backer/deamon/internal/api/v1/proto"
	"github.com/lmriccardo/backer/deamon/internal/core/service"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
)

type HandlerFunc = func(ctx *gin.Context, srv *service.Service)

func WrapHandler(srv *service.Service, fn HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) { fn(ctx, srv) }
}

func HandleListJobs(ctx *gin.Context, srv *service.Service) {

}

// @Summary Job Create Request
// @Description Request the registration of a new job
// @Tags jobs
// @Accept json
// @Produce json
// @Param job body proto.CreateJobRequest true "Job configuration"
// @Success 200 {object} proto.BaseResponse
// @Failure 400 {object} proto.BaseResponse "Validation Error"
// @Failure 409 {object} proto.BaseResponse "Duplicate Job name"
// @Failure 501 {object} proto.BaseResponse "Internal Server Error (most likely db fault)"
// @Router /v1/jobs/create [post]
func HandleJobCreateRequest(ctx *gin.Context, srv *service.Service) {
	// Binds the request body to the request structure and applies
	// defaults where necessary
	log.Println("[NEW REQUEST] Received new job creation request")

	var req proto.CreateJobRequest
	if err := utils.MustBindWithJSON(&req, ctx.Request); err != nil {
		ctx.JSON(http.StatusBadRequest, proto.NewErrorResponseFromErrs(err))
		return
	}

	// Validates the request payload using internal gin validator
	if ok, errs := ValidateRequest(req); !ok {
		ctx.JSON(http.StatusBadRequest, proto.NewErrorResponseFromErrs(errs...))
		return
	}

	// Create the job and
	if err := srv.CreateJob(ctx.Request.Context(), &req, nil); err != nil {
		// If the error returned by the function regards the job
		// with that name already existing, then we need to return 409
		status_code := proto.GetCodeFromError(err)
		ctx.JSON(status_code, proto.NewErrorResponseFromErrs(err))
		return
	}

	msg := fmt.Sprintf("Created job %v", req.Name)
	ctx.JSON(http.StatusCreated, proto.NewSuccessResponse(msg))
}

// @Summary Job Run Request
// @Description Request the execution of the job with given name
// @Tags jobs
// @Accept json
// @Produce json
// @Param job body proto.RunJobRequest true "Job execution, notification and logging configuration"
// @Success 202 {object} proto.RunJobResponse "Job execution accepted and queued"
// @Failure 404 {object} proto.BaseResponse "Job Not Found with given name"
// @Failure 501 {object} proto.BaseResponse "Generic Internal Server Error"
// @Router /v1/jobs/:name/run [post]
func HandleRunJobRequest(ctx *gin.Context, srv *service.Service) {
	// Take the job name from the api parameters
	job_name := ctx.Params.ByName("name")
	log.Printf("[NEW REQUEST] Received new job running request for: %s\n", job_name)

	// Bind the JSON payload from the request to the struct
	req := proto.RunJobRequest{Name: job_name}
	if err := utils.MustBindWithJSON(&req, ctx.Request); err != nil {
		ctx.JSON(http.StatusBadRequest, proto.NewErrorResponseFromErrs(err))
		return
	}

	// Request the service to enqueue the job to run
	run, err := srv.RunJob(ctx.Request.Context(), &req, nil)
	if err != nil {
		status_code := proto.GetCodeFromError(err)
		ctx.JSON(status_code, proto.NewErrorResponseFromErrs(err))
		return
	}

	ctx.JSON(http.StatusAccepted, proto.NewRunJobResponse(run))
}

// RegisterHandlers registers v1 handlers
func RegisterHandlers(rg *gin.RouterGroup, srv *service.Service) {
	_ = RegisterValidators() // Registers all validators

	grp := rg.Group("/v1")
	grp.Group("/jobs").
		GET("/", WrapHandler(srv, HandleListJobs)).
		POST("/create", WrapHandler(srv, HandleJobCreateRequest)).
		POST("/:name/run", WrapHandler(srv, HandleRunJobRequest))
}
