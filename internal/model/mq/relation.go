package mq

type FollowingCdcJson struct {
	CanalJson
	Data []FollowingCdc `json:"data"`
}

type FollowerCdcJson struct {
	CanalJson
	Data []FollowerCdc `json:"data"`
}

type FollowingCdc struct {
	Id          string `json:"id"`
	FollowerId  string `json:"follower_id"`
	Type        string `json:"type"`
	FollowingId string `json:"following_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type FollowerCdc struct {
	Id          string `json:"id"`
	FollowingId string `json:"following_id"`
	Type        string `json:"type"`
	FollowerId  string `json:"follower_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}
