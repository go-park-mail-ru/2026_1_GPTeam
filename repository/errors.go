package repository

import "fmt"

type ErrorFunc func(args ...interface{}) error

var NothingInTableError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("no rows in result set")
}
var InvalidDataInTableError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("unable to scan: invalid data in table") // ToDo: paste err in message
}
var DuplicatedDataError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("duplicated data in table: %v", args)
}
var ConstraintError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("constraint error: %v", args)
}
var TooManyRowsError ErrorFunc = func(args ...interface{}) error {
	return fmt.Errorf("too many rows in result set")
}
