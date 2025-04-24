package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/database"
	"fansX/internal/script/commentservicescript"
	"fansX/internal/util"
	syncx "fansX/pkg/sync"
	"log/slog"
	"strconv"
	"time"

	"fansX/services/comment/internal/svc"
	"fansX/services/comment/proto/commentRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReplyCommentListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReplyCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyCommentListLogic {
	return &GetReplyCommentListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReplyCommentListLogic) GetReplyCommentList(in *commentRpc.GetReplyCommentListReq) (*commentRpc.CommentListResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user get reply comment list", "ContentId", in.ContentId, "RootId", in.RootId, "limit", in.Limit, "Timestamp", in.TimeStamp)
	key := "ReplyCommentList:" + strconv.FormatInt(in.ContentId, 10) + ":" + strconv.FormatInt(in.RootId, 10)

	records, status := l.GetFromRedis(timeout, key, in.Limit, in.TimeStamp, logger)
	if status&StatusError != 0 {
		return l.HandleError(timeout, in, logger)
	} else if status&StatusNeedRebuild != 0 {
		return l.HandleRebuild(timeout, records, status, in, logger)
	} else if status&StatusFind != 0 {
		return l.HandleFind(timeout, records, status, in, logger)
	}
	return l.HandleNotFind(timeout, in, logger)

}

func (l *GetReplyCommentListLogic) HandleError(ctx context.Context, in *commentRpc.GetReplyCommentListReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	records, err := l.GetFromTiDB(ctx, in.ContentId, in.RootId, in.Limit, in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}
	return RecordsToResp(records), nil
}

func (l *GetReplyCommentListLogic) HandleRebuild(ctx context.Context, records []ListRecord, status int, in *commentRpc.GetReplyCommentListReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	go l.RebuildRedis(in, logger)
	if len(records) == int(in.Limit) {
		return RecordsToResp(records), nil
	}
	if status&StatusIsAll != 0 {
		return RecordsToResp(records), nil
	}
	logger.Info("reply comment list records in redis is not enough get other from tidb")
	records, err := l.GetFromTiDB(ctx, in.ContentId, in.RootId, in.Limit, in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}
	return RecordsToResp(records), nil
}

func (l *GetReplyCommentListLogic) HandleFind(ctx context.Context, records []ListRecord, status int, in *commentRpc.GetReplyCommentListReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	if len(records) == int(in.Limit) {
		return RecordsToResp(records), nil
	}
	if status&StatusIsAll != 0 {
		return RecordsToResp(records), nil
	}
	logger.Info("reply comment list records in redis is not enough get other from tidb")
	records, err := l.GetFromTiDB(ctx, in.ContentId, in.RootId, in.Limit, in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}

	return RecordsToResp(records), nil
}

func (l *GetReplyCommentListLogic) HandleNotFind(ctx context.Context, in *commentRpc.GetReplyCommentListReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	go l.RebuildRedis(in, logger)
	records, err := l.GetFromTiDB(ctx, in.ContentId, in.RootId, in.Limit, in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}

	return RecordsToResp(records), nil
}

func (l *GetReplyCommentListLogic) RebuildRedis(in *commentRpc.GetReplyCommentListReq, logger *slog.Logger) {
	key := "ReplyCommentList:" + strconv.FormatInt(in.ContentId, 10) + ":" + strconv.FormatInt(in.RootId, 10)
	mutex := l.svcCtx.Sync.NewMutex(key+":mutex", syncx.WithUtil(time.Second*5), syncx.WithTTL(time.Second))
	err := mutex.TryLock()
	if err == nil {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		records, err := l.GetFromTiDB(timeout, in.ContentId, in.RootId, 1001, time.Now().Add(time.Hour).UnixMilli(), logger)
		if err != nil {
			return
		}
		if len(records) > 1000 {
			l.BuildRedis(key, records, "false", logger)
		} else {
			l.BuildRedis(key, records, "true", logger)
		}
		_ = mutex.Unlock()
	}
	return
}

func (l *GetReplyCommentListLogic) BuildRedis(key string, records []ListRecord, all string, logger *slog.Logger) {
	data := make([]interface{}, len(records)*2+4)
	for i, v := range records {
		b, err := json.Marshal(records[i])
		if err != nil {
			logger.Error("marshal json to build reply comment redis:" + err.Error())
			return
		}
		data[i*2+1] = string(b)
		data[i*2] = v.CreatedAt
	}
	data[len(data)-4] = -1
	data[len(data)-3] = 5
	data[len(data)-2] = -2
	data[len(data)-1] = all
	err := l.svcCtx.Executor.Execute(context.Background(), commentservicescript.Build, []string{key, "60"}, data).Err()
	if err != nil {
		logger.Error("build reply comment list redis:" + err.Error())
	}
}

func (l *GetReplyCommentListLogic) GetFromRedis(ctx context.Context, key string, limit int64, timestamp int64, logger *slog.Logger) ([]ListRecord, int) {
	executor := l.svcCtx.Executor
	resp, err := executor.Execute(ctx, commentservicescript.GetByTime, []string{key}, limit, timestamp).Result()
	if err != nil {
		logger.Error("execute lua to get reply comment from redis:" + err.Error())
		return nil, StatusError
	}
	listInter := resp.([]interface{})
	status, _ := strconv.ParseInt(listInter[len(listInter)-1].(string), 10, 64)
	if status&StatusNotFind != 0 {
		logger.Debug("reply comment list not exists from redis")
		return nil, int(status)
	} else if status&StatusNeedRebuild != 0 {
		logger.Info("get reply comment list from redis but need to rebuild")
	} else {
		logger.Info("get reply comment list from redis")
	}
	res := make([]ListRecord, len(listInter)-1)
	for i := 0; i < len(listInter)-1; i++ {
		err = json.Unmarshal([]byte(listInter[i].(string)), &res[i])
		if err != nil {
			logger.Error("parse reply comment list from redis:" + err.Error())
			return nil, StatusError
		}
	}
	return res, int(status)
}

func (l *GetReplyCommentListLogic) GetFromTiDB(ctx context.Context, contentId int64, RootId int64, limit int64, timestamp int64, logger *slog.Logger) ([]ListRecord, error) {
	db := l.svcCtx.DB.WithContext(ctx)
	records := make([]database.Comment, 0)
	err := db.Where("content_id = ? and root_id = ? and status = ? and created_at <= ?", contentId, RootId, database.CommentStatusCommon, timestamp).
		Limit(int(limit)).Order("created_at desc").Find(records).Error
	if err != nil {
		logger.Error("get reply comment list from tidb:" + err.Error())
		return nil, err
	}
	logger.Info("get reply comment list from tidb")
	res := make([]ListRecord, len(records))
	for i, v := range records {
		res[i].LongTextUri = v.LongTextUri
		res[i].ShortText = v.ShortText
		res[i].CreatedAt = v.CreatedAt
		res[i].ContentId = v.ContentId
		res[i].RootId = v.RootId
		res[i].ParentId = v.ParentId
		res[i].UserId = v.UserId
		res[i].CommentId = v.Id
	}

	return res, nil
}
