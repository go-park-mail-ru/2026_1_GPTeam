package rate_limiter

import "fmt"

var NoIpInSavedError = fmt.Errorf("no ip saved")
var UnableToGetIp = fmt.Errorf("unable to get ip")
var WrongServerIpAddress = fmt.Errorf("wrong server ip")
var ResultNotOkError = fmt.Errorf("result not ok")
