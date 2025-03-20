package apperr

import (
	"errors"
)

// database error
var (
	ErrNoRows = errors.New("no rows in the result set")
)

// http error
var (
	ErrEmptyBody = errors.New("empty request body")
)
