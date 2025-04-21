package logic

import (
	"context"
	"encoding/binary"
	"errors"
	"fansX/internal/model/database"
	"fansX/internal/script/commentservicescript"
	"fansX/internal/util"
	"github.com/go-redsync/redsync/v4"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"strings"
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
	hot := cache.IsHotKey(key)

	count, status, err := l.GetCommentCountFromRedis(timeout, key, logger)
	if status == StatusFind {
		return &commentRpc.GetCommentCountResp{Count: count}, err
	}

	count, err = l.GetCommentCountFromTiDB(timeout, logger, in)
	if err != nil {
		return &commentRpc.GetCommentCountResp{Count: count}, err
	}
	l.BuildCommentCountCache(key, count, hot, status, logger)
	return &commentRpc.GetCommentCountResp{Count: count}, err
}

func (l *GetCommentCountLogic) GetCommentCountFromRedis(ctx context.Context, key string, logger *slog.Logger) (count int64, status int, err error) {
	executor := l.svcCtx.Executor
	res, err := executor.Execute(ctx, commentservicescript.GetCountScript, []string{key}).Result()
	if err != nil {
		logger.Error("get comment count from redis:" + err.Error())
		return 0, StatusError, err
	}
	if res == nil {
		return 0, StatusNotFind, nil
	}

	str := strings.Split(res.(string), ";")
	count, _ = strconv.ParseInt(str[0], 10, 64)
	stamp, _ := strconv.ParseInt(str[1], 10, 64)
	if time.Now().Unix() > stamp {
		logger.Debug("get comment count form redis and need to rebuild")
		return count, StatusNeedRebuild, nil
	}
	logger.Debug("get comment count from redis")
	return count, StatusFind, nil
}

func (l *GetCommentCountLogic) GetCommentCountFromTiDB(ctx context.Context, logger *slog.Logger, in *commentRpc.GetCommentCountReq) (count int64, err error) {

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

func (l *GetCommentCountLogic) BuildCommentCountCache(key string, count int64, hot bool, status int, logger *slog.Logger) {
	if hot {
		logger.Debug("build comment count local cache")
		l.BuildCommentCountLocal(key, count)
	}
	if status == StatusNeedRebuild || status == StatusNotFind {
		go func() {
			l.BuildCommentCountRedis(key, count, logger)
		}()
	}
}

func (l *GetCommentCountLogic) BuildCommentCountLocal(key string, count int64) {
	b := make([]byte, 8)
	binary.BigEndian.AppendUint64(b, uint64(count))
	if count >= 10000 {
		l.svcCtx.Cache.Set(key, b, 60)
	} else if count >= 1000 {
		l.svcCtx.Cache.Set(key, b, 10)
	} else {
		l.svcCtx.Cache.Set(key, b, 5)
	}
	return
}

func (l *GetCommentCountLogic) BuildCommentCountRedis(key string, count int64, logger *slog.Logger) {
	mutex := l.svcCtx.RedSync.NewMutex(key, redsync.WithExpiry(time.Second))
	err := mutex.TryLock()
	if err != nil {
		logger.Debug("try to get redis lock failed:" + err.Error())
		return
	}
	logger.Debug("try to get redis lock success")
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	value := strconv.FormatInt(count, 10) + ";" + strconv.FormatInt(time.Now().Add(time.Second*60).Unix(), 10)
	l.svcCtx.Client.Set(timeout, key, value, 70*time.Second)
	_, _ = mutex.Unlock()
	return
}

const (
	StatusError       = 0
	StatusFind        = 1
	StatusNeedRebuild = 2
	StatusNotFind     = 3
)
