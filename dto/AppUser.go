package dto

type AppUser struct {
	Id       int
	Hash     string
	Jwt      string
	Username string
	ImageUrl *string
	Instance *string
}

func NewAppUser(
	id int,
	hash string,
	jwt string,
	username string,
	image *string,
	instance *string,
) *AppUser {
	return &AppUser{
		Id:       id,
		Hash:     hash,
		Jwt:      jwt,
		Username: username,
		ImageUrl: image,
		Instance: instance,
	}
}
