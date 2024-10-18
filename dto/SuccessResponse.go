package dto

type SuccessResponse struct {
	Message  string `json:"message"`
	Posts    string `json:"posts"`
	Comments string `json:"comments"`
	Both     string `json:"both"`
}

func NewSuccessResponse(url string) *SuccessResponse {
	return &SuccessResponse{
		Message:  "Success! You can find your feed at " + url,
		Posts:    url,
		Comments: url + "?include=comments",
		Both:     url + "?include=posts,comments",
	}
}
