package trace

import (
	"context"

	"backend/utils/logs"
	"backend/utils/rand"
)

func InjectTraceID(ctx context.Context) context.Context {
	tace_id := rand.GenTraceID()
	return context.WithValue(ctx, logs.TraceIDContextKey, tace_id)
}

func InjectSpan(ctx context.Context) context.Context {
	// 获取当前的 span_id 作为 parent_span_id
	var parentSpanID string
	if currentSpanID := ctx.Value(logs.SpanIDContextKey); currentSpanID != nil {
		if spanIDStr, ok := currentSpanID.(string); ok {
			parentSpanID = spanIDStr
		}
	}

	// 设置新的 span_id
	ctx = context.WithValue(ctx, logs.SpanIDContextKey, rand.GenSpanID())

	// 如果有 parent_span_id，则设置它
	if parentSpanID != "" {
		ctx = context.WithValue(ctx, logs.ParentSpanIDContextKey, parentSpanID)
	}

	return ctx
}
