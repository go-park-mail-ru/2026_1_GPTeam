package context_helper

import "context"

type contextKey string

const (
	ContextKeyRequestId contextKey = "request_id"
	ContextKeyUser      contextKey = "user"
)

func GetRequestIdFromContext(ctx context.Context) string {
	requestId, ok := ctx.Value(ContextKeyRequestId).(string)
	if !ok {
		return ""
	}
	return requestId
}
