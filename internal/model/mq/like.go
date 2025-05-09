package mq

type LikeCountCdcJson struct {
	CanalJson
	Data []LikeCountCdc `json:"data"`
	Old  []LikeCountCdc `json:"old"`
}

type LikeCountCdc struct {
	Id       string `json:"id"`
	Business string `json:"business"`
	LikeId   string `json:"like_id"`
	Status   string `json:"status"`
	Count    string `json:"count"`
}

type LikeKafkaJson struct {
	TimeStamp int64 `json:"time_stamp"`
	Business  int32 `json:"business"`
	UserId    int64 `json:"user_id"`
	LikeId    int64 `json:"like_id"`
	Cancel    bool  `json:"cancel"`
}
