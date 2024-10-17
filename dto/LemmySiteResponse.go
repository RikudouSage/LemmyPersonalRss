package dto

type LemmySiteResponse struct {
	MyUser *LemmyMyUser `json:"my_user"`
}

type LemmyMyUser struct {
	LocalUserView LemmyLocalUserView `json:"local_user_view"`
}

type LemmyLocalUserView struct {
	Person LemmyPerson `json:"person"`
}
