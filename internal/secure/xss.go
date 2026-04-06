package secure

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

var once sync.Once
var strictPolicy *bluemonday.Policy

func XssSanitizerInit() {
	once.Do(func() {
		strictPolicy = bluemonday.StrictPolicy()
	})
}

func SanitizeXss(data string) string {
	if data == "" {
		return ""
	}
	return strictPolicy.Sanitize(data)
}
