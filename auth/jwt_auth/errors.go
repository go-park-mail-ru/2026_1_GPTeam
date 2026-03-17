package jwt_auth

import "fmt"

type ErrorFunc func(args ...interface{}) error

var WrongSigningMethodError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("unexpected signing method: %v\n", args)
}
var InvalidTokenId ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("invalid token id %v\n", args)
}
var JwtSecretError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("secret must be at least 8 bytes\n")
}
var JwtVersionError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("JWT_VERSION env variable not set\n")
}
