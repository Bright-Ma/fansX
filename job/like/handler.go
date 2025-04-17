package like

import (
	"encoding/json"
	"fansX/internal/model/database"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

type Handler struct {
	db         *gorm.DB
	client     *redis.Client
	nextWindow int64
}

func NewHandler(db *gorm.DB, client *redis.Client) (*Handler, error) {
	record := database.WindowLatest{}
	err := db.Take(&record).Error
	if err != nil {
		return nil, err
	}
	h := &Handler{
		db:         db,
		client:     client,
		nextWindow: record.NextWindow,
	}

	return h, nil
}

func (h *Handler) handle() {
	window := h.nextWindow
	db := h.db
	for {
		for window >= time.Now().Add(-time.Minute*30).Unix() {
			time.Sleep(time.Minute)
		}
		records := &[]database.TimeWindow{}
		err := db.Where("window = ?", window).Take(records).Error
		if err != nil {
			slog.Error("take window:"+err.Error(), "window", window)
			time.Sleep(time.Second)
			continue
		}
		check := make(map[[3]int64]int64)
		repair := make(map[[2]int64]int64)

		for _, record := range *records {
			body := database.WindowBody{}
			err = json.Unmarshal(record.Body, &body)
			if err != nil {
				panic(err.Error())
			}

			for like, user := range body.Like {
				var add int64
				if like[1] > 0 {
					add = 1
				} else {
					add = -1
				}
				for _, v := range user {
					check[[3]int64{like[0], like[1], v}] += add
				}
			}
		}

		for i, v := range check {
			if v > 1 {
				repair[[2]int64{i[0], i[1]}] -= v - 1
			} else if v < -1 {
				repair[[2]int64{i[0], i[1]}] -= v + 1

			}
		}

		tx := h.db.Begin()

		for i, v := range repair {
			if v != 0 {
				err = tx.Model(database.LikeCount{}).Where("business = ? and like_id = ?", i[0], i[1]).
					Update("count", gorm.Expr("count + ?", v)).Error
				if err != nil {
					slog.Error("repair like count:"+err.Error(), "business", i[0], "like_id", i[1], "repair", v)
					tx.Rollback()
					return
				}
			}
		}

		err = tx.Take(&database.WindowLatest{}).Update("next_window", window+1).Error
		if err != nil {
			slog.Error("update latest window:" + err.Error())
			tx.Rollback()
			return
		}
		tx.Commit()
		window++
	}

	return
}
