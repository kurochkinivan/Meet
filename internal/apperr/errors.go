package apperr

import (
	"errors"
)

// database error
var (
	ErrNoRows = errors.New("no rows in the result set")
)

// usecase error
var (
	ErrUserExists = errors.New("user with this phone already exists")
)

// http error
var (
	ErrEmptyBody     = errors.New("empty request body")
	ErrSerializeData = errors.New("failed to serialize/deserialize data")
)
