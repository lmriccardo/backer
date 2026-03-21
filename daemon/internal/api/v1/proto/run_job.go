package proto

import (
	"fmt"

	"github.com/lmriccardo/backer/deamon/internal/core/domain"
)

type RunJobRequest struct {
	Name string // Job name

	DryRun bool `json:"dry_run" binding:"required"` // Runs a dry-run or a true run
	Notify bool `json:"notify"  binding:"required"` // Enable/Disable notifications
	Log    bool `json:"log"     binding:"required"` // Enable/Disable logging
}

type RunJobResponse struct {
	BaseResponse // It is a base reponse

	RunId   string // The Id of the run
	JobName string // The name of the job
	Status  string // The current status of the run
}

func NewRunJobResponse(run *domain.Run) RunJobResponse {
	response := RunJobResponse{}
	response.BaseResponse.Message = fmt.Sprintf("Queued oneshot run for job %s", run.JobName)
	response.BaseResponse.Status = StatusSuccess
	response.RunId = run.Id
	response.JobName = run.JobName
	response.Status = run.Status.String()
	return response
}
