package cache

import (
	"context"
	"fmt"
	"sync"
)

var instanceCache sync.Map

func Get[R any](ctx context.Context, collection string, id int64) (R, bool) {
	key := fmt.Sprint("%s-%d", collection, id)
	var zero R
	ctxCache, ok := instanceCache.Load(ctx)
	if !ok {
		return zero, false
	}
	if v, ok := ctxCache.(*sync.Map).Load(key); ok {
		return v.(R), true
	} else {
		return zero, false
	}
}

func Store[R any](ctx context.Context, collection string, id int64, value R) {
	key := fmt.Sprint("%s-%d", collection, id)
	ctxCache, ok := instanceCache.Load(ctx)
	if !ok {
		ctxCache = &sync.Map{}
		instanceCache.Store(ctx, ctxCache)
		// Cleanup the cache when the context is done
		context.AfterFunc(ctx, func() { instanceCache.Delete(ctx) })
	}
	ctxCache.(*sync.Map).Store(key, value)
}
