package rate_limiter

import "fmt"

var UnableToGetIp = fmt.Errorf("unable to get ip")
var WrongServerIpAddress = fmt.Errorf("wrong server ip")
