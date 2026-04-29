package secure

import (
	"fmt"
	"net/http"
	"os"
)

func AddCSPHeader(w http.ResponseWriter) {
	defaultSrc := "'self'"
	scriptSrc := "'self'"
	styleSrc := "'self' 'unsafe-inline'"
	imageSrc := "'self' data:"
	fontSrc := "'self' data:"
	connectSrc := "'self'"
	frameSrc := "'self'"
	frameAncestors := "'self'"

	adUrl := os.Getenv("ADVERTISEMENT_URL")
	if adUrl != "" {
		frameSrc = fmt.Sprintf("'self' %s", adUrl)
	}

	value := fmt.Sprintf(
		"default-src %s; script-src %s; style-src %s; img-src %s; font-src %s; connect-src %s; frame-src %s; frame-ancestors %s;",
		defaultSrc, scriptSrc, styleSrc, imageSrc, fontSrc, connectSrc, frameSrc, frameAncestors,
	)
	w.Header().Set("Content-Security-Policy", value)
}
