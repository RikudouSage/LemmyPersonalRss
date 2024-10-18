package user

import (
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/dto"
	"LemmyPersonalRss/helper"
	"LemmyPersonalRss/lemmy"
	"fmt"
	"net/http"
)

var api *lemmy.Api

func findJwt(request *http.Request) *string {
	var jwt string
	for _, cookie := range request.Cookies() {
		if cookie.Name != "jwt" {
			continue
		}

		jwt = cookie.Value
		break
	}

	if jwt == "" {
		return nil
	}

	return &jwt
}

func findLemmyUser(request *http.Request) *dto.LemmyPerson {
	jwt := findJwt(request)
	if jwt == nil {
		return nil
	}

	return api.UserByJwt(*jwt, nil)
}

func GetCurrentFromHttpContext(request *http.Request, db database.Database) *dto.AppUser {
	lemmyUser := findLemmyUser(request)
	if lemmyUser == nil {
		return nil
	}

	user := db.FindByUserId(lemmyUser.Id)

	if user == nil {
		return nil
	}

	return user
}

func UpdateUserData(appUser *dto.AppUser, db database.Database) error {
	lemmyUser := api.UserByJwt(appUser.Jwt, appUser.Instance)
	appUser.ImageUrl = lemmyUser.Avatar
	if appUser.Instance == nil {
		appUser.Instance = &config.GlobalConfiguration.Instance
	}

	err := db.StoreUser(appUser)
	if err != nil {
		fmt.Println(err)
	}

	return err
}

func CreateFromHttpContext(request *http.Request, db database.Database) *dto.AppUser {
	lemmyUser := findLemmyUser(request)
	if lemmyUser == nil {
		return nil
	}

	if lemmyUser.Banned {
		fmt.Println("User is banned")
		return nil
	}
	if lemmyUser.Deleted {
		fmt.Println("User is deleted")
		return nil
	}

	secureHash, err := helper.RandomString(32)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	user := &dto.AppUser{
		Id:       lemmyUser.Id,
		Hash:     secureHash,
		Jwt:      *findJwt(request),
		Username: lemmyUser.Name,
		ImageUrl: lemmyUser.Avatar,
		Instance: &config.GlobalConfiguration.Instance,
	}
	if lemmyUser.Avatar != nil && *lemmyUser.Avatar != "" {
		user.ImageUrl = lemmyUser.Avatar
	}

	err = db.StoreUser(user)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return user
}

func init() {
	api = &lemmy.Api{}
}
