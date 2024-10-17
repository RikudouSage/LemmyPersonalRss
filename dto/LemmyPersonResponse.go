package dto

type LemmyPersonResponse struct {
	Posts []LemmyPostView `json:"posts"`
}

type LemmyPostView struct {
	Post    LemmyPost   `json:"post"`
	Creator LemmyPerson `json:"creator"`
}

type LemmyPost struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Url           *string   `json:"url"`
	Body          *string   `json:"body"`
	CreatorId     int       `json:"creator_id"`
	CommunityId   int       `json:"community_id"`
	Removed       bool      `json:"removed"`
	Locked        bool      `json:"locked"`
	Published     DateTime  `json:"published"`
	Updated       *DateTime `json:"updated,omitempty"`
	Deleted       bool      `json:"deleted"`
	Nsfw          bool      `json:"nsfw"`
	ThumbnailUrl  *string   `json:"thumbnail_url,omitempty"`
	ActivityPubId string    `json:"ap_id"`
	Local         bool      `json:"local"`
}

type LemmyPerson struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	DisplayName *string  `json:"display_name,omitempty"`
	Banned      bool     `json:"banned"`
	Published   DateTime `json:"published"`
	ActorId     string   `json:"actor_id"`
	Local       bool     `json:"local"`
	Deleted     bool     `json:"deleted"`
	BotAccount  bool     `json:"bot_account"`
	Avatar      *string  `json:"avatar,omitempty"`
	Banner      *string  `json:"banner,omitempty"`
	Bio         *string  `json:"bio,omitempty"`
}
