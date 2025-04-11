package public

import (
	"bilibili/internal/model/database"
	"bilibili/internal/model/mq"
	"github.com/IBM/sarama"
	"strconv"
)

package main

import (
"bilibili/internal/model/database"
"bilibili/internal/model/mq"
"context"
"encoding/json"
"errors"
"github.com/IBM/sarama"
"gorm.io/gorm"
"gorm.io/gorm/clause"
"log/slog"
"strconv"
"time"
)

type Handler struct {
	db *gorm.DB
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.MetaContentJson{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}
		if message.IsDdl {
			continue
		}
		slog.Info("consume kafka message", "dml_type", message.Type)
		id, _ := strconv.ParseInt(message.Data[0].Id, 10, 64)
		version, _ := strconv.ParseInt(message.Data[0].Version, 10, 64)
		status, _ := strconv.ParseInt(message.Data[0].Status, 10, 64)
		oldStatus, _ := strconv.ParseInt(message.Data[0].OldStatus, 10, 64)
		userid, _ := strconv.ParseInt(message.Data[0].UserId, 10, 64)
		record := &database.InvisibleContentInfo{
			Id:              id,
			Version:         version,
			Status:          int(status),
			OldStatus:       int(oldStatus),
			Desc:            message.Data[0].Desc,
			Userid:          userid,
			Title:           message.Data[0].Title,
			PhotoUriList:    message.Data[0].PhotoUriList,
			ShortText:       message.Data[0].ShortText,
			LongTextUri:     message.Data[0].LongTextUri,
			VideoUriList:    message.Data[0].VideoUriList,
			OldPhotoUriList: message.Data[0].OldPhotoUriList,
			OldShortText:    message.Data[0].OldShortTextUri,
			OldLongTextUri:  message.Data[0].OldLongTextUri,
			OldVideoUriList: message.Data[0].OldVideoUriList,
		}
		slog.Info("Handle record change", "id", record.Id, "userId", record.Userid, "status", record.Status, "version", record.Version)
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
		if record.Status == database.ContentStatusCheck {
			err = h.HandleCheck(timeout, record)
		} else if record.Status == database.ContentStatusPass {
			err = h.HandlePass(timeout, record)
		} else if record.Status == database.ContentStatusDelete {
			err = h.HandleDelete(timeout, record)
		} else {
			err = h.HandleNotPass(timeout, record)
		}
		if err == nil {
			session.MarkMessage(msg, "")
		}

		cancel()
	}

	return nil
}

func (h *Handler) HandleCheck(ctx context.Context, record *database.InvisibleContentInfo) error {

	/*
		检查内容是否合规，此处省略
		正常做法：kafka->consumer(审核)->kafka->consumer(状态更改)
	*/

	record.OldStatus = record.Status
	record.Version++
	record.Status = database.ContentStatusPass
	record.OldPhotoUriList = record.PhotoUriList
	record.OldShortText = record.ShortText
	record.OldLongTextUri = record.LongTextUri
	record.OldVideoUriList = record.VideoUriList

	db := h.db.WithContext(ctx)
	latest := &database.InvisibleContentInfo{}

	tx := db.Begin()
	err := tx.Select("version").Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(latest, record.Id).Error
	if err != nil {
		tx.Commit()
		slog.Error("take record:" + err.Error())
		return err
	}

	if latest.Version == record.Version-1 {

		err = tx.Take(latest, record.Id).Updates(record).Error
		if err != nil {
			tx.Rollback()
			slog.Error("update record" + err.Error())
			return nil
		}

		tx.Commit()
		return nil
	}

	tx.Commit()
	slog.Warn("consume message is not latest", "version", record.Version-1, "latest", latest.Version)
	return nil

}

func (h *Handler) HandlePass(ctx context.Context, record *database.InvisibleContentInfo) error {
	latest := &database.VisibleContentInfo{}

	tx := h.db.WithContext(ctx).Begin()

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(latest, record.Id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("take visible content info record:" + err.Error())
		tx.Rollback()
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Info("record not found form visible table")
	} else {
		if latest.Version >= record.Version {
			tx.Rollback()
			slog.Warn("consume message version to small", "version", record.Version, "visible_version", latest.Version)
			return nil
		}
	}

	latest = &database.VisibleContentInfo{
		Id:           record.Id,
		Version:      record.Version,
		Userid:       record.Userid,
		Status:       record.Status,
		Title:        record.Title,
		PhotoUriList: record.PhotoUriList,
		ShortText:    record.ShortText,
		LongTextUri:  record.LongTextUri,
		VideoUriList: record.VideoUriList,
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Create(latest).Error

		if err != nil {
			tx.Rollback()
			slog.Error("insert to visible content info table:" + err.Error())
		} else {
			tx.Commit()
		}
		return err

	} else {

		err = tx.Take(&database.VisibleContentInfo{}, latest.Id).Updates(latest).Error
		if err != nil {
			tx.Rollback()
			slog.Error("update visible content info record:" + err.Error())
		} else {
			tx.Commit()
		}

		return err

	}

}

func (h *Handler) HandleDelete(ctx context.Context, record *database.InvisibleContentInfo) error {
	db := h.db.WithContext(ctx)

	err := db.Take(&database.VisibleContentInfo{}, record.Id).
		Update("status", database.ContentStatusDelete).Update("version", record.Version).Error
	if err != nil {
		slog.Error("set visible content info table failed:" + err.Error())
		return err
	}

	return nil
}

// HandleNotPass 通常不会走到这一步
func (h *Handler) HandleNotPass(ctx context.Context, record *database.InvisibleContentInfo) error {
	return nil

}
