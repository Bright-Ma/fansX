package mq

type FollowingCanalJson struct {
	CanalJson
	Data []Following `json:"data"`
}

type FollowerCanalJson struct {
	CanalJson
	Data []Follower `json:"data"`
}

type Following struct {
	Id          string `json:"id"`
	FollowerId  string `json:"follower_id"`
	Type        string `json:"type"`
	FollowingId string `json:"following_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}

type Follower struct {
	Id          string `json:"id"`
	FollowingId string `json:"following_id"`
	Type        string `json:"type"`
	FollowerId  string `json:"follower_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}
