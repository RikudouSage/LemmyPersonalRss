package controller

import (
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/helper/response"
	"LemmyPersonalRss/user"
	"fmt"
	"net/http"
	"strings"
)

func HandleInit(writer http.ResponseWriter, request *http.Request, feedUrl string, db database.Database) {
	if config.GlobalConfiguration.Logging {
		fmt.Println("GET /rss/init called")
		defer func() {
			fmt.Println("GET /rss/init finished")
		}()
	}

	if config.GlobalConfiguration.Instance == "" {
		if config.GlobalConfiguration.Logging {
			fmt.Println("Called init, but no instance is configured")
		}

		err := response.WriteForbiddenResponse(
			map[string]string{
				"error": "Automatic init is not enabled",
			},
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	currentUser := user.GetCurrentFromHttpContext(request, db)
	if currentUser == nil {
		currentUser = user.CreateFromHttpContext(request, db)
	}

	if currentUser == nil {
		if config.GlobalConfiguration.Logging {
			fmt.Println("User is not logged in")
		}
		err := response.WriteUnauthorizedResponse(
			map[string]string{
				"error": "Failed to get current user. Are you logged in?",
			},
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	url := strings.Replace(feedUrl, "{hash}", currentUser.Hash, -1)

	err := response.WriteOkResponse(
		map[string]string{
			"message": "Success! You can find your feed at " + url,
			"url":     url,
		},
		writer,
	)
	if err != nil {
		fmt.Println(err)
	}
}
