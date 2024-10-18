package cache

import (
	"fmt"
	"time"
)

func RunUnlessCached(cachePool ItemPool, cacheKey string, duration *time.Duration, callable func()) {
	cacheItem := cachePool.Get(cacheKey)
	if cacheItem.Hit() {
		return
	}

	callable()

	cacheItem.Set(true)
	cacheItem.SetExpiresAfter(duration)
	err := cachePool.Store(cacheItem)
	if err != nil {
		fmt.Println(err)
	}
}
