package mq

type UserLoginKafkaJson struct {
	Id int64 `json:"id"`
}

type UserCdcJson struct {
	CanalJson
	Data []UserCdc `json:"data"`
	Old  []UserCdc `json:"old"`
}

type UserCdc struct {
	Id       string `json:"id"`
	Phone    string `json:"phone"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
