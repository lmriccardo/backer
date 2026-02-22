package v1

import "github.com/lmriccardo/backer/deamon/internal/platform/utils"

type StatusType int

const (
	StatusSuccess StatusType = iota
	StatusFailure
)

// BaseResponse type for all api calls (can be embedded into specific response)
type BaseResponse struct {
	Status  StatusType `json:"status"`  // Response status
	Message string     `json:"message"` // Message filled up only if success
	Errors  []string   `json:"errors"`  // Errors filled up only if failure
}

func NewErrorResponse(errs ...string) BaseResponse {
	return BaseResponse{Status: StatusFailure, Errors: errs}
}

func NewErrorResponseFromErrs(errs ...error) BaseResponse {
	return NewErrorResponse(utils.Map(
		func(err error) string { return err.Error() },
		errs,
	)...)
}

func NewSuccessResponse(msg string) BaseResponse {
	return BaseResponse{Status: StatusSuccess, Message: msg}
}
