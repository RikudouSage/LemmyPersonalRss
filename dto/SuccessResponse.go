package dto

type SuccessResponse struct {
	Message string `json:"message"`
	Url     string `json:"url"`
}

func NewSuccessResponse(url string) *SuccessResponse {
	return &SuccessResponse{
		Message: "Success! You can find your feed at " + url,
		Url:     url,
	}
}
