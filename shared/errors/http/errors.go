package http

import (
	"github.com/pkg/errors"
)

var (
	// ErrBadRequestBody when we can't decode the request body.
	ErrBadRequestBody = errors.New("bad request body")
)
