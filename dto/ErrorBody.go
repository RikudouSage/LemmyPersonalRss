package dto

type ErrorBody struct {
	Error string `json:"error"`
}

func NewErrorBody(error string) *ErrorBody {
	return &ErrorBody{Error: error}
}
