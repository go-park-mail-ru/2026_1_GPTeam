package context_helper

import "context"

func GetRequestIdFromContext(ctx context.Context) string {
	requestId, ok := ctx.Value("request_id").(string)
	if !ok {
		return ""
	}
	return requestId
}
