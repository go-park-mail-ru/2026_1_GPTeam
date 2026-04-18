package secure

import (
	"fmt"
	"net/http"
)

func AddCspHeader(w http.ResponseWriter) {
	defaultSrc := "'self'"
	scriptSrc := "'self'"
	styleSrc := "'self' 'unsafe-inline'"
	imageSrc := "'self' data:"
	fontSrc := "'self' data:"
	connectSrc := "'self'"
	value := fmt.Sprintf("default-src %s; script-src %s; style-src %s; img-src %s; font-src %s; connect-src %s;",
		defaultSrc, scriptSrc, styleSrc, imageSrc, fontSrc, connectSrc)
	w.Header().Add("Content-Security-Policy", value)
}
