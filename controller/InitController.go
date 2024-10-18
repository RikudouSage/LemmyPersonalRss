package controller

import (
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/dto"
	"LemmyPersonalRss/helper"
	"LemmyPersonalRss/helper/response"
	"LemmyPersonalRss/lemmy"
	"LemmyPersonalRss/user"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HandleInit(writer http.ResponseWriter, request *http.Request, feedUrl string, db database.Database) {
	defer request.Body.Close()

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
			dto.NewErrorBody("Automatic init is not enabled"),
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
			dto.NewErrorBody("Failed to get current user. Are you logged in?"),
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	url := strings.Replace(feedUrl, "{hash}", currentUser.Hash, -1)

	err := response.WriteOkResponse(
		dto.NewSuccessResponse(url),
		writer,
	)
	if err != nil {
		fmt.Println(err)
	}
}

func HandleRegister(writer http.ResponseWriter, request *http.Request, api *lemmy.Api, db database.Database, feedUrl string) {
	defer request.Body.Close()

	var instance string
	if config.GlobalConfiguration.Instance == "" {
		instance = request.Host
	} else {
		instance = config.GlobalConfiguration.Instance
	}

	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)

		err = response.WriteInternalErrorResponse(
			dto.NewErrorBody("Failed to read body, please try again later."),
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	var body dto.RegisterRequest
	err = helper.MapJson(rawBody, &body)
	if err != nil {
		fmt.Println(err)

		err = response.WriteBadRequestResponse(
			dto.NewErrorBody("Failed to parse body, make sure your request is correct."),
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	if body.Instance == nil || *body.Instance == "" {
		body.Instance = &instance
	}

	if body.Jwt == "" {
		err := response.WriteBadRequestResponse(
			dto.NewErrorBody("You must provide a JWT token."),
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	lemmyUser := api.UserByJwt(body.Jwt, body.Instance)
	if lemmyUser == nil {
		err := response.WriteBadRequestResponse(
			dto.NewErrorBody("The JWT token or instance is invalid."),
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}

		return
	}

	appUser := db.FindByUserId(lemmyUser.Id)
	if appUser == nil {
		appUser = user.CreateFromLemmyUser(lemmyUser, db, body.Jwt, body.Instance)
	}

	url := strings.Replace(feedUrl, "{hash}", appUser.Hash, -1)
	url = strings.Replace(feedUrl, "{instance}", *appUser.Instance, -1)

	err = response.WriteOkResponse(
		dto.NewSuccessResponse(url),
		writer,
	)
	if err != nil {
		fmt.Println(err)
	}
}
