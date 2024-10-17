package cache

import "time"

type internalItem struct {
	value      any
	validUntil *time.Time
}

type InMemoryCacheItemPool struct {
	items map[string]*internalItem
}

func (receiver *InMemoryCacheItemPool) Get(key string) Item {
	item, ok := receiver.items[key]
	hit := true

	if !ok {
		item = &internalItem{
			value:      nil,
			validUntil: nil,
		}
		hit = false
	} else if item.validUntil != nil && time.Now().After(*item.validUntil) {
		delete(receiver.items, key)
		hit = false
		item = &internalItem{
			value:      nil,
			validUntil: nil,
		}
	}

	return &DefaultItem{
		value:      item.value,
		key:        key,
		validUntil: item.validUntil,
		hit:        hit,
	}
}

func (receiver *InMemoryCacheItemPool) Store(item Item) error {
	if receiver.items == nil {
		receiver.items = make(map[string]*internalItem)
	}
	receiver.items[item.Key()] = &internalItem{
		value:      item.Get(),
		validUntil: item.ExpiresAt(),
	}

	return nil
}
