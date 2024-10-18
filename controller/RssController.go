package controller

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/feed"
	"LemmyPersonalRss/helper"
	"LemmyPersonalRss/lemmy"
	"LemmyPersonalRss/user"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HandleRssFeed(
	writer http.ResponseWriter,
	request *http.Request,
	feedPath string,
	db database.Database,
	cachePool cache.ItemPool,
	api *lemmy.Api,
) {
	urlHash := request.PathValue("hash")
	if config.GlobalConfiguration.Logging {
		fmt.Println(strings.Replace("GET /"+feedPath+" called", "{hash}", urlHash, -1))
		defer func() {
			fmt.Println(strings.Replace("GET /"+feedPath+" finished", "{hash}", urlHash, -1))
		}()
	}

	writer.Header().Set("Content-Type", "application/json")

	appUser := db.FindByHash(urlHash)

	if appUser == nil {
		response := map[string]string{
			"error": "The RSS feed could not be found.",
		}
		raw, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
		}

		writer.WriteHeader(http.StatusNotFound)
		_, err = writer.Write(raw)
		if err != nil {
			fmt.Println(err)
		}

		if config.GlobalConfiguration.Logging {
			fmt.Println("RSS feed not found")
		}
		return
	}

	go helper.RunUnlessCached(
		cachePool,
		fmt.Sprintf("%s.%d", "user_refresh_blocker", appUser.Id),
		&config.GlobalConfiguration.CacheDuration,
		func() {
			err := user.UpdateUserData(appUser, db)
			if err != nil {
				fmt.Println(err)
			}
		},
	)

	page := helper.GetQueryStringInt(request, "page", 1)
	perPage := helper.GetQueryStringInt(request, "limit", 20)

	saved := api.GetSavedPosts(appUser, page, perPage)

	rssFeed := feed.CreateFeedFromPosts(saved, appUser, page, cachePool)

	rss, err := rssFeed.ToRss()
	if err != nil {
		fmt.Println(err)

		response := map[string]string{
			"error": "The RSS feed could not be generated, please try again later.",
		}
		raw, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
		}

		writer.WriteHeader(http.StatusInternalServerError)
		_, err = writer.Write(raw)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	writer.Header().Set("Content-Type", "application/rss+xml")
	_, err = writer.Write([]byte(rss))
	if err != nil {
		fmt.Println(err)
		return
	}
}
