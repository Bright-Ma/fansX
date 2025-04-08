package mq

import "bilibili/internal/model/database"

type MetaToPublicMessage struct {
	Type int                           `json:"type"`
	Data database.InvisibleContentInfo `json:"data"`
}

type MetaToMetaMessage struct {
	ContentId int64 `json:"content_id"`
}

var (
	MetaToPublicTypeDel = 1
	MetaToPublicTypeIns = 2
	MetaToPublicTypeUpd = 3
)
