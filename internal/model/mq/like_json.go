package mq

type LikeCountJson struct {
	CanalJson
	Data []LikeCount `json:"data"`
	Old  []LikeCount `json:"old"`
}

type LikeCount struct {
	Id       string `json:"id"`
	Business string `json:"business"`
	LikeId   string `json:"like_id"`
	Status   string `json:"status"`
	Count    string `json:"count"`
}
