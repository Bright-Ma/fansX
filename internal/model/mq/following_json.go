package mq

type FollowingCanalJson struct {
	CanalJson
	Data []Following `json:"data"`
}

type RetryFollowingCanalJson struct {
	Type string      `json:"type"`
	Data []Following `json:"data"`
}

type Following struct {
	Id          string `json:"id"`
	FollowerId  string `json:"follower_id"`
	Type        string `json:"type"`
	FollowingId string `json:"following_id"`
	UpdatedAt   string `json:"updated_at"`
	CreatedAt   string `json:"created_at"`
}
