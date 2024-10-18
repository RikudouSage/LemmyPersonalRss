package controller

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/feed"
	"LemmyPersonalRss/helper"
	"LemmyPersonalRss/helper/response"
	"LemmyPersonalRss/lemmy"
	"LemmyPersonalRss/user"
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
	defer request.Body.Close()

	urlHash := request.PathValue("hash")
	if config.GlobalConfiguration.Logging {
		fmt.Println(strings.Replace("GET /"+feedPath+" called", "{hash}", urlHash, -1))
		defer func() {
			fmt.Println(strings.Replace("GET /"+feedPath+" finished", "{hash}", urlHash, -1))
		}()
	}

	appUser := db.FindByHash(urlHash)

	if appUser == nil {
		if config.GlobalConfiguration.Logging {
			fmt.Println("RSS feed not found")
		}

		err := response.WriteNotFoundResponse(
			map[string]string{
				"error": "The RSS feed could not be found.",
			},
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	go cache.RunUnlessCached(
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
		err := response.WriteInternalErrorResponse(
			map[string]string{
				"error": "The RSS feed could not be generated, please try again later.",
			},
			writer,
		)
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
