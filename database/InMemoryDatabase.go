package database

import (
	"LemmyPersonalRss/dto"
)

type InMemoryDatabase struct {
	users []*dto.AppUser
}

func (receiver *InMemoryDatabase) FindByHash(userHash string) *dto.AppUser {
	for _, user := range receiver.users {
		if user.Hash == userHash {
			return user
		}
	}

	return nil
}

func (receiver *InMemoryDatabase) StoreUser(user *dto.AppUser) error {
	receiver.users = append(receiver.users, user)
	return nil
}

func (receiver *InMemoryDatabase) FindByUserId(userId int) *dto.AppUser {
	for _, user := range receiver.users {
		if user.Id == userId {
			return user
		}
	}

	return nil
}
