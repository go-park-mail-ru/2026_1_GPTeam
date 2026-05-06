package groq

import "errors"

var (
	ErrInternalClient      = errors.New("internal client error")
	ErrClientRateLimit     = errors.New("client rate limit exceeded")
	ErrClientInvalidFile   = errors.New("invalid file format")
	ErrClientUnauthorized  = errors.New("unauthorized access")
	ErrClientEmptyResult   = errors.New("empty result from API")
)
