package secure

import "fmt"

var CsrfSecretError = fmt.Errorf("CSRF secret error")
var InvalidCsrfError = fmt.Errorf("CSRF invalid")
var InvalidCsrfSignatureError = fmt.Errorf("CSRF invalid signature")
