package mq

type Like struct {
	TimeStamp int64 `json:"time_stamp"`
	Business  int32 `json:"business"`
	UserId    int64 `json:"user_id"`
	LikeId    int64 `json:"like_id"`
	Cancel    bool  `json:"cancel"`
}
