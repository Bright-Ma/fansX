package main

import (
	"context"
	"errors"
	"fansX/internal/model/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log/slog"
)

func NewFactory() StrategyFactory {
	return StrategyFactory{strategySet: make(map[int]Strategy)}
}

func (s *StrategyFactory) GetStrategy(status int) Strategy {
	return s.strategySet[status]
}

func (s *StrategyFactory) RegisterStrategy(status int, strategy Strategy) {
	s.strategySet[status] = strategy
}

func (s *CheckStrategy) Handle(ctx context.Context, db *gorm.DB, record *database.InvisibleContentInfo) error {
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

	db = db.WithContext(ctx)
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

func (s *PassStrategy) Handle(ctx context.Context, db *gorm.DB, record *database.InvisibleContentInfo) error {
	latest := &database.VisibleContentInfo{}

	tx := db.WithContext(ctx).Begin()

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(latest, record.Id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("take visible content info record:" + err.Error())
		tx.Rollback()
		return err
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("take visible content info record:" + err.Error())
		tx.Commit()
		return err
	} else if err == nil {
		if latest.Version >= record.Version {
			tx.Commit()
			slog.Warn("message version is too small", "version", record.Version, "latest_version", latest.Version)
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
	if err != nil {
		err = tx.Create(latest).Error
		if err != nil {
			tx.Rollback()
			slog.Error("create visible content info:" + err.Error())
			return err
		}
		tx.Commit()
		return nil
	} else {
		err = tx.Take(&database.VisibleContentInfo{}, latest.Id).Updates(latest).Error
		if err != nil {
			tx.Rollback()
			slog.Error("update visible content info:" + err.Error())
			return err
		}
		tx.Commit()
		return nil
	}
}

func (s *DeleteStrategy) Handle(ctx context.Context, db *gorm.DB, record *database.InvisibleContentInfo) error {
	db = db.WithContext(ctx)

	err := db.Take(&database.VisibleContentInfo{}, record.Id).
		Update("status", database.ContentStatusDelete).Update("version", record.Version).Error
	if err != nil {
		slog.Error("set visible status to delete:" + err.Error())
		return err
	}

	return nil
}

func (s *NotPassStrategy) Handle(ctx context.Context, db *gorm.DB, record *database.InvisibleContentInfo) error {
	// 通常不会走到这一步
	return nil
}
