package rate_limiter

import "fmt"

var ErrUnableToGetIp = fmt.Errorf("unable to get ip")
var ErrWrongServerIpAddress = fmt.Errorf("wrong server ip")
