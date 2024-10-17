package cache

import (
	. "LemmyPersonalRss/helper"
	"time"
)

type DefaultItem struct {
	value      any
	key        string
	validUntil *time.Time
	hit        bool
}

func (receiver *DefaultItem) Key() string {
	return receiver.key
}

func (receiver *DefaultItem) Hit() bool {
	return receiver.hit
}

func (receiver *DefaultItem) Get() any {
	return receiver.value
}

func (receiver *DefaultItem) Set(value any) {
	receiver.value = value
}

func (receiver *DefaultItem) ExpiresAt() *time.Time {
	return receiver.validUntil
}

func (receiver *DefaultItem) SetExpiresAt(expiresAt *time.Time) {
	receiver.validUntil = expiresAt
}

func (receiver *DefaultItem) SetExpiresAfter(expiresAfter *time.Duration) {
	if expiresAfter != nil {
		receiver.validUntil = ToPointer(time.Now().Add(*expiresAfter))
	}
}
