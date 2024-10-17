package database

import (
	"LemmyPersonalRss/dto"
)

type InMemoryDatabase struct {
	users map[int]*dto.AppUser
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
	if receiver.users == nil {
		receiver.users = make(map[int]*dto.AppUser)
	}
	receiver.users[user.Id] = user
	return nil
}

func (receiver *InMemoryDatabase) FindByUserId(userId int) *dto.AppUser {
	user, ok := receiver.users[userId]
	if !ok {
		return nil
	}

	return user
}
