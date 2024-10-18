package helper

import (
	"fmt"
	"net/http"
	"strconv"
)

func GetQueryStringInt(request *http.Request, name string, defaultValue int) int {
	raw := request.URL.Query().Get(name)
	if raw == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		fmt.Println(err)
		return defaultValue
	}

	return parsed
}
