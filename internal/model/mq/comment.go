package mq

type CommentCountCdcJson struct {
	CanalJson
	Data []CommentCountCdc `json:"data"`
	Old  []CommentCountCdc `json:"old"`
}

type CommentCountCdc struct {
	Id       string `json:"id"`
	Business string `json:"business"`
	CountId  string `json:"count_id"`
	Count    string `json:"count"`
}

type CommentKafkaJson struct {
	Id          int64  `json:"id"`
	UserId      int64  `json:"user_id"`
	ContentId   int64  `json:"content_id"`
	RootId      int64  `json:"root_id"`
	ParentId    int64  `json:"parent_id"`
	ShortText   string `json:"short_text"`
	LongTextUri string `json:"long_text_uri"`
}

type DelCommentKafkaJson struct {
	UserId    int64 `json:"user_id"`
	CommentId int64 `json:"comment_id"`
}
