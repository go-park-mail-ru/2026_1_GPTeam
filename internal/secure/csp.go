package secure

import (
	"fmt"
	"net/http"
)

func AddCspHeader(w http.ResponseWriter) {
	defaultSrc := "'self'"
	scriptSrc := "'self'"
	styleSrc := "'self'"
	imageSrc := "'self' data:"
	fontSrc := "'self' data:"
	value := fmt.Sprintf("default-src %s; script-src %s; style-src %s; img-src %s; font-src %s;", defaultSrc, scriptSrc, styleSrc, imageSrc, fontSrc)
	w.Header().Add("Content-Security-Policy", value)
}
