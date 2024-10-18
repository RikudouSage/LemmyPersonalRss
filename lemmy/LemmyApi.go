package lemmy

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/dto"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Api struct {
	Cache cache.ItemPool
}

func (receiver *Api) UserByJwt(jwt string, instance *string) *dto.LemmyPerson {
	if instance == nil {
		instance = &config.GlobalConfiguration.Instance
	}

	httpClient := &http.Client{}
	request, err := http.NewRequest("GET", "https://"+*instance+"/api/v3/site", nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	request.Header.Add("Authorization", "Bearer "+jwt)
	response, err := httpClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer response.Body.Close()
	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var siteResponse dto.LemmySiteResponse
	err = json.Unmarshal(rawBody, &siteResponse)
	if err != nil {
		fmt.Println(err)
	}

	if siteResponse.MyUser == nil {
		fmt.Println("Invalid JWT, no user found")
		return nil
	}

	return &siteResponse.MyUser.LocalUserView.Person
}

func (receiver *Api) getSavedStuffResponse(user *dto.AppUser, page int, perPage int) (result *dto.LemmyPersonResponse) {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}

	var cacheItem cache.Item
	if receiver.Cache != nil {
		key := fmt.Sprintf("%s.%d.%d.%d", "saved_profile", user.Id, page, perPage)
		cacheItem = receiver.Cache.Get(key)
	} else {
		cacheItem = &cache.DefaultItem{}
	}

	if cacheItem.Hit() {
		return cacheItem.Get().(*dto.LemmyPersonResponse)
	}

	var instance *string
	if user.Instance != nil {
		instance = user.Instance
	} else {
		instance = &config.GlobalConfiguration.Instance
	}

	url := fmt.Sprintf(
		"https://%s/api/v3/user?username=%s&sort=New&saved_only=true&page=%d&limit=%d",
		*instance,
		user.Username,
		page,
		perPage,
	)

	httpClient := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	request.Header.Add("Authorization", "Bearer "+user.Jwt)
	response, err := httpClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer response.Body.Close()
	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var personResponse dto.LemmyPersonResponse
	err = json.Unmarshal(rawBody, &personResponse)
	if err != nil {
		fmt.Println(err)
	}

	cacheItem.Set(&personResponse)
	cacheItem.SetExpiresAfter(&config.GlobalConfiguration.CacheDuration)

	if receiver.Cache != nil {
		err = receiver.Cache.Store(cacheItem)
		if err != nil {
			fmt.Println(err)
		}
	}

	return &personResponse
}

func (receiver *Api) GetSavedPosts(user *dto.AppUser, page int, perPage int) (result []*dto.LemmyPostView) {
	result = make([]*dto.LemmyPostView, 0, perPage)
	personResponse := receiver.getSavedStuffResponse(user, page, perPage)
	if personResponse == nil {
		return
	}

	for _, post := range personResponse.Posts {
		result = append(result, &post)
	}

	return
}
