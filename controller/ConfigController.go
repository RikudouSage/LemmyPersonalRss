package controller

import (
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/dto"
	"LemmyPersonalRss/helper/response"
	"fmt"
	"net/http"
)

func HandleConfigEndpoint(writer http.ResponseWriter) {
	if !config.GlobalConfiguration.EnableConfigEndpoint {
		err := response.WriteNotFoundResponse(
			dto.NewErrorBody("Not found"),
			writer,
		)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	err := response.WriteOkResponse(
		map[string]any{
			"instance":           config.GlobalConfiguration.Instance,
			"singleInstanceMode": config.GlobalConfiguration.SingleInstanceMode,
		},
		writer,
	)
	if err != nil {
		fmt.Println(err)
	}
}
