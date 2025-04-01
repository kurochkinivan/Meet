package apperr

import (
	"errors"
)

// database error
var (
	ErrNoRows = errors.New("no rows in the result set")
)

// service error
var (
	ErrUserExists = errors.New("user with this phone already exists")
)

// transport error
var (
	ErrEmptyBody     = errors.New("empty request body")
	ErrSerializeData = errors.New("failed to serialize/deserialize data")
)
