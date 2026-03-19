package jwt_auth

import (
	"errors"
)

var WrongSigningMethodError = errors.New("unexpected signing method")
var InvalidTokenId = errors.New("invalid token id")
var JwtSecretError = errors.New("secret must be at least 8 bytes")
var JwtVersionError = errors.New("JWT_VERSION env variable not set")
