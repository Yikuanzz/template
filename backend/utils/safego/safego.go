package safego

import (
	"context"
	"fmt"
	"runtime/debug"

	"backend/utils/logs"
)

func Recovery(ctx context.Context) {
	e := recover()
	if e == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	err := fmt.Errorf("%v", e)
	logs.CtxErrorf(ctx, "[catch panic] err = %v \n stacktrace:\n%s", err, debug.Stack())
}

func Go(ctx context.Context, fn func()) {
	go func() {
		defer Recovery(ctx)
		fn()
	}()
}
