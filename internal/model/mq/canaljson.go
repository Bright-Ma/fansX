package mq

type CanalJson struct {
	Id       int64    `json:"id"`
	Database string   `json:"database"`
	Table    string   `json:"table"`
	PkNames  []string `json:"pkNames"`
	IsDdl    bool     `json:"isDdl"`
	Type     string   `json:"type"`
	Es       int64    `json:"es"`
	Ts       int64    `json:"ts"`
	Sql      string   `json:"sql"`
}
