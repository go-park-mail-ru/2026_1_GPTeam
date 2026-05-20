package secure

import "fmt"

var ErrCsrfSecret = fmt.Errorf("CSRF secret error")
var ErrInvalidCsrf = fmt.Errorf("CSRF invalid")
var ErrInvalidCsrfSignature = fmt.Errorf("CSRF invalid signature")
