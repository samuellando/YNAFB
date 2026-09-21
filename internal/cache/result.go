package cache

import (
	"context"
	"sync"
)

var resultCache sync.Map

func Result[R any](ctx context.Context, key string, f func() (R, error)) (R, error) {
	ctxCache, ok := resultCache.Load(ctx)
	if !ok {
		ctxCache = &sync.Map{}
		resultCache.Store(ctx, ctxCache)
		// Cleanup the cache when the context is done
		context.AfterFunc(ctx, func() { resultCache.Delete(ctx) })
	}
	if v, ok := ctxCache.(*sync.Map).Load(key); ok {
		return v.(R), nil
	}
	// Cache miss
	v, err := f()
	if err != nil {
		var zero R
		return zero, err
	}
	ctxCache.(*sync.Map).Store(key, v)
	return v, nil
}

func InvalidateResults(ctx context.Context) {
	resultCache.Delete(ctx)
}
