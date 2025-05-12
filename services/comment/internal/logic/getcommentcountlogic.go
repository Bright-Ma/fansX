package logic

import (
	"context"
	"encoding/binary"
	"errors"
	"fansX/internal/model/database"
	"fansX/internal/util"
	syncx "fansX/pkg/sync"
	"fansX/services/comment/internal/script"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"time"

	"fansX/services/comment/internal/svc"
	"fansX/services/comment/proto/commentRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentCountLogic {
	return &GetCommentCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommentCountLogic) GetCommentCount(in *commentRpc.GetCommentCountReq) (*commentRpc.GetCommentCountResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user get comment count", "business", in.Business, "countId", in.CountId)
	cache := l.svcCtx.Cache

	key := "CommentCount:" + strconv.Itoa(int(in.Business)) + ":" + strconv.FormatInt(in.CountId, 10)

	b, ok := cache.Get(key)
	if ok {
		logger.Info("get comment count from local cache")
		count := binary.BigEndian.Uint64(b)
		return &commentRpc.GetCommentCountResp{Count: int64(count)}, nil
	}
	count, status := l.GetFromRedis(timeout, key, logger)
	if status == StatusFind {
		return &commentRpc.GetCommentCountResp{Count: count}, nil
	} else if status == StatusNeedRebuild {
		go func() {
			mutex := l.svcCtx.Sync.NewMutex(key+":mutex", syncx.WithUtil(time.Second*5), syncx.WithTTL(time.Second))
			if mutex.TryLock() == nil {
				count, err := l.GetFromTiDB(timeout, logger, in)
				if err == nil {
					l.BuildRedis(key, count, logger)
				}
				_ = mutex.Unlock()
			}
		}()
		return &commentRpc.GetCommentCountResp{Count: count}, nil
	} else if status == StatusError {
		count, err := l.GetFromTiDB(timeout, logger, in)
		if err != nil {
			return nil, err
		}
		return &commentRpc.GetCommentCountResp{Count: count}, nil
	}
	count, err := l.GetFromTiDB(timeout, logger, in)
	if err != nil {
		return nil, err
	}
	go l.BuildRedis(key, count, logger)
	return &commentRpc.GetCommentCountResp{Count: count}, nil
}

func (l *GetCommentCountLogic) GetFromRedis(ctx context.Context, key string, logger *slog.Logger) (int64, int) {
	executor := l.svcCtx.Executor
	resp, err := executor.Execute(ctx, script.GetCountScript, []string{key}).Result()
	if err != nil {
		logger.Error("get comment count from redis:" + err.Error())
		return 0, StatusError
	}
	status, _ := strconv.ParseInt(resp.([]interface{})[1].(string), 10, 64)
	if status == StatusNotFind {
		logger.Debug("comment count not exists from redis")
		return 0, int(status)
	} else if status == StatusNeedRebuild {
		logger.Info("get comment count from redis but need to rebuild")
	} else {
		logger.Info("get comment list from redis")
	}

	count, _ := strconv.ParseInt(resp.([]interface{})[0].(string), 10, 64)
	return count, int(status)
}

func (l *GetCommentCountLogic) GetFromTiDB(ctx context.Context, logger *slog.Logger, in *commentRpc.GetCommentCountReq) (count int64, err error) {

	db := l.svcCtx.DB.WithContext(ctx)
	record := database.CommentCount{}

	err = db.Where("business = ? and count_id = ?", in.Business, in.CountId).Take(record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("get comment count from tidb:" + err.Error())
		return 0, err
	} else if err != nil {
		logger.Info("comment count not exists in tidb")
		return 0, err
	}

	logger.Debug("get comment count from tidb")
	count = record.Count
	return
}

func (l *GetCommentCountLogic) BuildRedis(key string, count int64, logger *slog.Logger) {
	logger.Debug("try to get redis lock success")
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	value := strconv.FormatInt(count, 10) + ";" + strconv.FormatInt(5, 10)
	l.svcCtx.Client.Set(timeout, key, value, 30*time.Second)
	return
}

const (
	StatusError       = 1 << 0
	StatusFind        = 1 << 1
	StatusNeedRebuild = 1 << 2
	StatusNotFind     = 1 << 3
	StatusIsAll       = 1 << 4
	StatusNotAll      = 1 << 5
)
