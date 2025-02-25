package apperr

import (
	"errors"
	"net/http"
)

type statusError struct {
	error
	status int
}

func (e *statusError) UnWrap() error   { return e.error }
func (e *statusError) HTTPStatus() int { return e.status }

func WithHTTPStatus(err error, status int) error {
	return &statusError{
		error:  err,
		status: status,
	}
}

func HTTPStatus(err error) int {
	if err == nil {
		return 0
	}

	var statusErr interface {
		error
		HTTPStatus() int
	}
	if errors.As(err, &statusErr) {
		return statusErr.HTTPStatus()
	}

	return http.StatusInternalServerError
}
