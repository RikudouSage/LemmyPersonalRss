package dto

type RegisterRequest struct {
	Jwt      string  `json:"jwt"`
	Instance *string `json:"instance,omitempty"`
}
