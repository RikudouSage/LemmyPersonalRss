package cache

import "time"

type Item interface {
	Key() string
	Hit() bool
	Get() any
	Set(value any)
	ExpiresAt() *time.Time
	SetExpiresAt(expiresAt *time.Time)
	SetExpiresAfter(expiresAfter *time.Duration)
}

type ItemPool interface {
	Get(key string) Item
	Store(item Item) error
	GetCleaner() Cleaner
}
