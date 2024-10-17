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

func (receiver *Api) UserByJwt(jwt string) *dto.LemmyPerson {
	httpClient := &http.Client{}
	request, err := http.NewRequest("GET", "https://"+config.GlobalConfiguration.Instance+"/api/v3/site", nil)
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

func (receiver *Api) GetSavedPosts(user *dto.AppUser, page int, perPage int) (result []*dto.LemmyPostView) {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}
	result = make([]*dto.LemmyPostView, 0, perPage)

	var cacheItem cache.Item

	if receiver.Cache != nil {
		key := fmt.Sprintf("%s.%d.%d.%d", "saved", user.Id, page, perPage)
		cacheItem = receiver.Cache.Get(key)
	} else {
		cacheItem = &cache.DefaultItem{}
	}

	if cacheItem.Hit() {
		return cacheItem.Get().([]*dto.LemmyPostView)
	}

	url := fmt.Sprintf(
		"https://%s/api/v3/user?username=%s&sort=New&saved_only=true&page=%d&limit=%d",
		config.GlobalConfiguration.Instance,
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

	for _, post := range personResponse.Posts {
		result = append(result, &post)
	}

	cacheItem.Set(result)
	cacheItem.SetExpiresAfter(&config.GlobalConfiguration.CacheDuration)

	if receiver.Cache != nil {
		err = receiver.Cache.Store(cacheItem)
		if err != nil {
			fmt.Println(err)
		}
	}

	return
}
