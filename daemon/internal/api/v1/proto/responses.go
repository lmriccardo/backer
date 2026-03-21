package proto

import (
	"net/http"

	"github.com/lmriccardo/backer/deamon/internal/core/errors"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
)

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

// GetCodeFromError maps custom service errors to known http
// status for responses. On default or unknown errors,
// status bad request is used.
func GetCodeFromError(err error) int {
	switch err.(type) {
	case *errors.DuplicateJobNameError:
		return http.StatusConflict
	case *errors.InvalidJobNameError:
		return http.StatusNotFound
	case *errors.ConfigurationError:
		return http.StatusBadRequest
	case *errors.DatabaseError:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}
