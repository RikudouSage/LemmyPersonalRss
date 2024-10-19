package cache

import "time"

type InMemoryCacheCleaner struct {
	cache *InMemoryCacheItemPool
}

func (receiver *InMemoryCacheCleaner) Clean() {
	now := time.Now()

	for key, item := range receiver.cache.items {
		if item.validUntil.Before(now) {
			delete(receiver.cache.items, key)
		}
	}
}
