package mq

type CommentCountJson struct {
	CanalJson
	Data []CommentCount     `json:"data"`
	Old  []CommentCountJson `json:"old"`
}

type CommentCount struct {
	Id       string `json:"id"`
	Business string `json:"business"`
	CountId  string `json:"count_id"`
	Count    string `json:"count"`
}
