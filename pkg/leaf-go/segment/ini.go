package segment

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)
import "gorm.io/driver/mysql"

func NewCreator(config *Config) (*Creator, error) {
	dsn := config.UserName + ":" + config.Password + "@" + "tcp(" + config.Address + ")" + "/IDCreator?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&IdTable{}); err != nil {
		return nil, err
	}

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx := db.WithContext(timeout).Begin()
	// 开启事务并且使用select for update 锁住记录，保证并发安全
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("Tag = ?", config.Name).First(&IdTable{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err = tx.Create(&IdTable{Tag: config.Name}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if err != nil {
		tx.Commit()
		return nil, err
	}

	err = tx.Model(&IdTable{}).Where("Tag = ?", config.Name).Update("MaxId", gorm.Expr("max_id + step")).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	record := &IdTable{}
	err = tx.Where("Tag = ?", config.Name).First(record).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	creator := &Creator{
		id: record.ID,
		db: db,
		ch: make(chan int, 1),
		old: &buffer{
			nextId: record.MaxId - record.Step + 1,
			maxId:  record.MaxId,
		},
		new: nil,
	}

	creator.old.preIndex = creator.old.nextId + record.Step/10
	go creator.preApplication()
	return creator, nil
}
