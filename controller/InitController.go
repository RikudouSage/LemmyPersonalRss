package controller

import (
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/user"
	"encoding/json"
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

	writer.Header().Set("Content-Type", "application/json")

	currentUser := user.GetCurrentFromHttpContext(request, db)
	if currentUser == nil {
		currentUser = user.CreateFromHttpContext(request, db)
	}

	if currentUser == nil {
		response := map[string]string{
			"error": "Failed to get current user. Are you logged in?",
		}
		raw, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
		}

		writer.WriteHeader(http.StatusUnauthorized)
		_, err = writer.Write(raw)
		if err != nil {
			fmt.Println(err)
		}

		if config.GlobalConfiguration.Logging {
			fmt.Println("User is not logged in")
		}
		return
	}

	url := strings.Replace(feedUrl, "{hash}", currentUser.Hash, -1)
	response := map[string]string{
		"message": "Success! You can find your feed at " + url,
		"url":     url,
	}

	raw, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}
	_, err = writer.Write(raw)
	if err != nil {
		fmt.Println(err)
	}
}
