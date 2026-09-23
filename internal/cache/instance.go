package cache

import (
	"context"
	"fmt"
	"sync"
)

var instanceCache sync.Map

func Get[R any](ctx context.Context, id int64, missFunc func() (R, error)) (R, error) {
	var zero R
	key := fmt.Sprintf("%T-%d", zero, id)
	ctxCache, ok := instanceCache.Load(ctx)
	if ok {
		if v, ok := ctxCache.(*sync.Map).Load(key); ok {
			return v.(R), nil
		}
	}
	res, err := missFunc()
	if err != nil {
		return zero, err
	}
	store(ctx, id, res)
	return res, nil
}

func Delete[R any](ctx context.Context, id int64) {
	var zero R
	key := fmt.Sprintf("%T-%d", zero, id)
	ctxCache, ok := instanceCache.Load(ctx)
	if !ok {
		return
	}
	ctxCache.(*sync.Map).Delete(key)
}

func store[R any](ctx context.Context, id int64, value R) {
	var zero R
	key := fmt.Sprintf("%T-%d", zero, id)
	ctxCache, ok := instanceCache.Load(ctx)
	if !ok {
		ctxCache = &sync.Map{}
		instanceCache.Store(ctx, ctxCache)
		// Cleanup the cache when the context is done
		context.AfterFunc(ctx, func() { instanceCache.Delete(ctx) })
	}
	ctxCache.(*sync.Map).Store(key, value)
}
