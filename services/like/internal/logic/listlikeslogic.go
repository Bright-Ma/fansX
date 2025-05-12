package logic

import (
	"context"
	"errors"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"fansX/services/like/internal/script"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"time"

	"fansX/services/like/internal/svc"
	"fansX/services/like/proto/likeRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListLikesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListLikesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListLikesLogic {
	return &ListLikesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListLikesLogic) ListLikes(in *likeRpc.ListLikesReq) (*likeRpc.ListLikesResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	group := l.svcCtx.Group
	db := l.svcCtx.DB.WithContext(timeout)

	key := "LikeList:" + strconv.Itoa(int(in.BusinessId)) + ":" + strconv.FormatInt(in.LikeId, 10)

	list, all, err := SearchLikeListFromRedis(timeout, logger, l.svcCtx, key, in)
	if err == nil {
		return SearchLikeListJudge(logger, db, in, list, all)
	}
	var rebuild bool
	if errors.Is(err, redis.Nil) {
		rebuild = true
	}

	res, err := group.Do(key, func() (interface{}, error) {
		res, all, err := SearchLikeListFromTiDB(db, int(in.BusinessId), in.LikeId, in.Limit, in.TimeStamp)
		if err == nil && rebuild {
			go RebuildLikeList(l.svcCtx.Executor, key, res, all)
		}
		return res, err
	})
	if err != nil {
		logger.Error("search user like list from tidb:" + err.Error())
		return nil, err
	}
	return &likeRpc.ListLikesResp{UserId: res.([][]int64)[0], TimeStamp: res.([][]int64)[1]}, nil
}

func SearchLikeListFromRedis(ctx context.Context, logger *slog.Logger, svc *svc.ServiceContext, key string, in *likeRpc.ListLikesReq) ([][]int64, bool, error) {
	executor := svc.Executor
	inter, err := executor.Execute(ctx, script.List, []string{key}, in.TimeStamp, in.Limit).Result()
	if err != nil {
		logger.Error("search like list from redis:" + err.Error())
		return nil, false, err
	}
	if inter == nil {
		logger.Info("not search like list from redis")
		return nil, false, redis.Nil
	}
	slice := inter.([]interface{})
	res := make([][]int64, 2)
	res[0] = make([]int64, 0)
	res[1] = make([]int64, 0)

	var all bool
	for i := 0; i < len(slice); i++ {
		if i == len(slice)-1 {
			if slice[i].(string) == "true" {
				all = true
			}
			break
		}
		res[i%2] = append(res[i%2], slice[i].(int64))
	}

	return res, all, nil
}

func RebuildLikeList(executor *lua.Executor, key string, list [][]int64, all bool) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	data := make([]interface{}, len(list[1])*2)
	for i := 0; i < len(data); i += 2 {
		data[i] = list[1][i/2]
		data[i+1] = list[0][i/2]
	}
	var a string
	if all {
		a = "true"
	} else {
		a = "false"
	}
	executor.Execute(timeout, script.BuildList, []string{key, a, "false", "60"}, data...)
}

func SearchLikeListJudge(logger *slog.Logger, db *gorm.DB, in *likeRpc.ListLikesReq, list [][]int64, all bool) (*likeRpc.ListLikesResp, error) {
	if int64(len(list[0])) < in.Limit {
		if all == true {
			return &likeRpc.ListLikesResp{UserId: list[0], TimeStamp: list[1]}, nil
		} else {
			res, _, err := SearchLikeListFromTiDB(db, int(in.BusinessId), in.LikeId, in.Limit, in.TimeStamp)
			if err != nil {
				logger.Error("search user like list from tidb:" + err.Error())
				return nil, err
			}
			return &likeRpc.ListLikesResp{
				UserId:    append(list[0], res[0]...),
				TimeStamp: append(list[1], res[1]...),
			}, nil
		}
	}
	return &likeRpc.ListLikesResp{UserId: list[0], TimeStamp: list[1]}, nil
}

func SearchLikeListFromTiDB(db *gorm.DB, business int, likeId int64, limit int64, timeStamp int64) ([][]int64, bool, error) {
	records := make([]database.Like, limit)
	err := db.Select("updated_at", "user_id").
		Where("business = ? and status = ? and like_id = ? and updated_at <= ?", business, database.LikeStatusLike, likeId, timeStamp).
		Find(&records).Limit(int(limit + 1)).Order("updated_at DESC").Error
	if err != nil {
		return nil, false, err
	}
	res := make([][]int64, 2)
	res[0] = make([]int64, max(len(records)-1, 0))
	res[1] = make([]int64, max(len(records)-1, 0))

	for i := range res[0] {
		res[0][i] = records[i].LikeId
		res[1][i] = records[i].UserId
	}

	if len(records) > int(limit) {
		return res, false, nil
	} else {
		return res, true, nil
	}
}
