package database

import "LemmyPersonalRss/dto"

type Database interface {
	FindByUserId(userId int) *dto.AppUser
	StoreUser(user *dto.AppUser) error
	FindByHash(userHash string) *dto.AppUser
}
