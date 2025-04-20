package mq

type LikeJson struct {
	CanalJson
	Data []struct {
		Id        string `json:"id"`
		Business  string `json:"business"`
		Status    string `json:"status"`
		UserId    string `json:"user_id"`
		UpdatedAt string `json:"updated_at"`
		LikeId    string `json:"like_id"`
		CreatedAt string `json:"created_at"`
	} `json:"data"`
}
