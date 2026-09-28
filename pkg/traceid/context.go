package traceid

import "context"

type ctxKey struct{}

func ContextWith(parent context.Context, traceID string) context.Context {
	return context.WithValue(parent, ctxKey{}, traceID)
}


func ExtractFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(ctxKey{}).(string)
	return traceID, ok
}
