package clients

import "errors"

var (
	ErrInternalClient     = errors.New("internal client error")
	ErrClientRateLimit    = errors.New("rate limit exceeded")
	ErrClientInvalidFile  = errors.New("invalid file format")
	ErrClientUnauthorized = errors.New("unauthorized request")
	ErrClientEmptyResult  = errors.New("empty result returned")
)
