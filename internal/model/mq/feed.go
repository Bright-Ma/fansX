package mq

type FeedListKafkaJson struct {
	List    []int64 `json:"list"`
	OldList []int64 `json:"old_list"`
}
