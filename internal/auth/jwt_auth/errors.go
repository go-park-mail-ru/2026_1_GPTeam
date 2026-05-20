package jwt_auth

import (
	"errors"
)

var ErrWrongSigningMethod = errors.New("unexpected signing method")
var ErrJwtSecret = errors.New("secret must be at least 8 bytes")
var ErrJwtVersion = errors.New("JWT_VERSION env variable not set")
