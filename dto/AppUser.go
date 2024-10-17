package dto

type AppUser struct {
	Id       int
	Hash     string
	Jwt      string
	Username string
	ImageUrl *string
}
