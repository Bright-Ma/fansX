package mq

type FeedListJson struct {
	List    []int64 `json:"list"`
	OldList []int64 `json:"old_list"`
}
