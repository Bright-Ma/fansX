package main

import (
	"context"
	"encoding/json"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"log/slog"
	"strconv"
	"time"
)

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	slog.Info("handler set up ")
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.MetaContentCdcJson{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}
		if message.IsDdl {
			session.MarkMessage(msg, "")
			continue
		}
		slog.Info("consume kafka message", "dml_type", message.Type)
		record := GetRecord(&message.Data[0])

		slog.Info("Handle record change", "id", record.Id, "userId", record.Userid, "status", record.Status, "version", record.Version)

		s := h.factory.GetStrategy(record.Status)

		if s != nil {
			// 无限重试
			_ = retry.Do(func() error {
				timeout, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.Handle(timeout, h.db, record)
				if err == nil {
					session.MarkMessage(msg, "")
				}
				return err
			}, retry.Attempts(1000), retry.DelayType(retry.BackOffDelay), retry.MaxDelay(time.Second))
		} else {
			slog.Error("get strategy failed", "status", record.Status)
		}

	}

	return nil
}

func GetRecord(data *mq.MetaContent) *database.InvisibleContentInfo {
	id, _ := strconv.ParseInt(data.Id, 10, 64)
	version, _ := strconv.ParseInt(data.Version, 10, 64)
	status, _ := strconv.ParseInt(data.Status, 10, 64)
	oldStatus, _ := strconv.ParseInt(data.OldStatus, 10, 64)
	userid, _ := strconv.ParseInt(data.UserId, 10, 64)
	record := &database.InvisibleContentInfo{
		Id:              id,
		Version:         version,
		Status:          int(status),
		OldStatus:       int(oldStatus),
		Desc:            data.Desc,
		Userid:          userid,
		Title:           data.Title,
		PhotoUriList:    data.PhotoUriList,
		ShortText:       data.ShortText,
		LongTextUri:     data.LongTextUri,
		VideoUriList:    data.VideoUriList,
		OldPhotoUriList: data.OldPhotoUriList,
		OldShortText:    data.OldShortTextUri,
		OldLongTextUri:  data.OldLongTextUri,
		OldVideoUriList: data.OldVideoUriList,
	}
	return record

}
