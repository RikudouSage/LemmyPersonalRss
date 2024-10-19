package controller

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/dto"
	"LemmyPersonalRss/feed"
	"LemmyPersonalRss/helper"
	"LemmyPersonalRss/helper/response"
	"LemmyPersonalRss/lemmy"
	"LemmyPersonalRss/user"
	"fmt"
	"github.com/gorilla/feeds"
	"net/http"
	"slices"
	"strings"
)

func HandleRssFeedEndpoint(
	writer http.ResponseWriter,
	request *http.Request,
	feedPath string,
	db database.Database,
	cachePool cache.ItemPool,
	api *lemmy.Api,
) {
	defer request.Body.Close()

	includeRaw := request.URL.Query().Get("include")
	if includeRaw == "" {
		includeRaw = "posts"
	}
	include := strings.Split(includeRaw, ",")

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
			dto.NewErrorBody("The RSS feed could not be found."),
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

	rssFeed := feed.CreateFeedForUser(appUser, page, cachePool)

	if slices.Contains(include, "posts") {
		savedPosts := api.GetSavedPosts(appUser, page, perPage)
		feed.AddPostsToFeed(rssFeed, savedPosts, cachePool, appUser)
	}
	if slices.Contains(include, "comments") {
		savedComments := api.GetSavedComments(appUser, page, perPage)
		feed.AddCommentsToFeed(rssFeed, savedComments, appUser)
	}

	slices.SortFunc(rssFeed.Items, func(a, b *feeds.Item) int {
		dateA := a.Created
		dateB := b.Created

		if dateA.Equal(dateB) {
			return 0
		}

		if dateA.After(dateB) {
			return -1
		}

		return 1
	})

	rss, err := rssFeed.ToRss()
	if err != nil {
		fmt.Println(err)
		err := response.WriteInternalErrorResponse(
			dto.NewErrorBody("The RSS feed could not be generated, please try again later."),
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
