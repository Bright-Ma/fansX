package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/database"
	"fansX/internal/script/commentservicescript"
	"fansX/internal/util"
	syncx "fansX/pkg/sync"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"fansX/services/comment/internal/svc"
	"fansX/services/comment/proto/commentRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentListByTimeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentListByTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListByTimeLogic {
	return &GetCommentListByTimeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type ListRecord struct {
	CommentId   int64  `json:"comment_id"`
	UserId      int64  `json:"user_id"`
	ContentId   int64  `json:"content_id"`
	RootId      int64  `json:"root_id"`
	ParentId    int64  `json:"parent_id"`
	CreatedAt   int64  `json:"created_at"`
	ShortText   string `json:"short_text"`
	LongTextUri string `json:"long_text_uri"`
}

func (l *GetCommentListByTimeLogic) GetCommentListByTime(in *commentRpc.GetCommentListByTimeReq) (*commentRpc.CommentListResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user get comment list by time", "content_id", in.ContentId, "limit", in.Limit, "time_stamp", in.TimeStamp)
	cache := l.svcCtx.Cache

	key := "CommentListByTime:" + strconv.FormatInt(in.ContentId, 10)

	v, ok := cache.Get(key)
	if ok {
		logger.Info("get comment list by time from local cache")
		value := &commentRpc.CommentListResp{}
		err := json.Unmarshal(v, value)
		if err != nil {
			panic(err.Error())
		}
		resp := RespGet(value, int(in.Limit), in.TimeStamp)
		if len(resp.CommentId) == int(in.Limit) {
			return resp, nil
		}
		logger.Info("comment list by time from local cache is not enough")
	}
	records, status := l.GetFromRedis(timeout, key, int(in.Limit), in.TimeStamp, logger)
	if status&StatusError != 0 {
		return l.HandleError(timeout, in.ContentId, in.Limit, in.TimeStamp, logger)
	} else if status&StatusNeedRebuild != 0 {
		return l.HandleRebuild(timeout, records, status, in, logger)
	} else if status&StatusFind != 0 {
		return l.HandleFind(timeout, records, status, in, logger)
	}
	return l.HandleNotFind(timeout, in, logger)
}

func (l *GetCommentListByTimeLogic) HandleError(ctx context.Context, contentId int64, limit int64, timestamp int64, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	records, err := l.GetFromTiDB(ctx, contentId, int(limit), timestamp, logger)
	if err != nil {
		return nil, err
	}
	return RecordsToResp(records), nil
}

func (l *GetCommentListByTimeLogic) HandleRebuild(ctx context.Context, records []ListRecord, status int, in *commentRpc.GetCommentListByTimeReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	go l.RebuildRedis(in.ContentId, logger)
	if len(records) == int(in.Limit) {
		return RecordsToResp(records), nil
	}
	if status&StatusIsAll != 0 {
		return RecordsToResp(records), nil
	}
	logger.Info("comment list by time records in redis is not enough get other from tidb")
	records, err := l.GetFromTiDB(ctx, in.ContentId, int(in.Limit), in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}
	return RecordsToResp(records), nil
}

func (l *GetCommentListByTimeLogic) HandleFind(ctx context.Context, records []ListRecord, status int, in *commentRpc.GetCommentListByTimeReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	if len(records) == int(in.Limit) {
		return RecordsToResp(records), nil
	}
	if status&StatusIsAll != 0 {
		return RecordsToResp(records), nil
	}
	logger.Info("comment list by time records in redis is not enough get other from tidb")
	records, err := l.GetFromTiDB(ctx, in.ContentId, int(in.Limit), in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}

	return RecordsToResp(records), nil
}

func (l *GetCommentListByTimeLogic) HandleNotFind(ctx context.Context, in *commentRpc.GetCommentListByTimeReq, logger *slog.Logger) (*commentRpc.CommentListResp, error) {
	go l.RebuildRedis(in.ContentId, logger)
	records, err := l.GetFromTiDB(ctx, in.ContentId, int(in.Limit), in.TimeStamp, logger)
	if err != nil {
		return nil, err
	}
	return RecordsToResp(records), nil
}

func (l *GetCommentListByTimeLogic) RebuildRedis(contentId int64, logger *slog.Logger) {
	mutex := l.svcCtx.Sync.NewMutex("CommentListByTime:"+strconv.FormatInt(contentId, 10)+":mutex", syncx.WithTTL(time.Second), syncx.WithUtil(time.Second*5))
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	err := mutex.TryLock()

	if err == nil {
		records, err := l.GetFromTiDB(timeout, contentId, 1001, time.Now().Add(time.Hour).UnixMilli(), logger)
		if err != nil {
			return
		}
		if len(records) > 1000 {
			l.BuildRedis("CommentListByTime:"+strconv.FormatInt(contentId, 10), records, "false")
		} else {
			l.BuildRedis("CommentListByTime:"+strconv.FormatInt(contentId, 10), records, "true")
		}
		_ = mutex.Unlock()
	}
	return
}

func (l *GetCommentListByTimeLogic) BuildRedis(key string, records []ListRecord, all string) {
	data := make([]interface{}, len(records)*2+4)
	for i, v := range records {
		b, err := json.Marshal(records[i])
		if err != nil {
			panic(err.Error())
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
		slog.Error("build comment list by time redis:" + err.Error())
	}
}

func (l *GetCommentListByTimeLogic) GetFromRedis(ctx context.Context, key string, limit int, timestamp int64, logger *slog.Logger) ([]ListRecord, int) {
	executor := l.svcCtx.Executor
	resp, err := executor.Execute(ctx, commentservicescript.GetByTime, []string{key}, limit, timestamp).Result()
	if err != nil {
		logger.Error("execute lua to get time list from redis:" + err.Error())
		return nil, StatusError
	}

	listInter := resp.([]interface{})
	status, _ := strconv.ParseInt(listInter[len(listInter)-1].(string), 10, 64)
	if status&StatusNotFind != 0 {
		logger.Debug("time comment list not exists from redis")
		return nil, int(status)
	} else if status&StatusNeedRebuild != 0 {
		logger.Info("get time comment list from redis but need to rebuild")
	} else {
		logger.Info("get time comment list from redis")
	}
	res := make([]ListRecord, len(listInter)-1)
	for i := 0; i < len(listInter)-1; i++ {
		err = json.Unmarshal([]byte(listInter[i].(string)), &res[i])
		if err != nil {
			panic(err.Error())
		}
	}
	return res, int(status)
}

func (l *GetCommentListByTimeLogic) GetFromTiDB(ctx context.Context, contentId int64, limit int, timestamp int64, logger *slog.Logger) ([]ListRecord, error) {
	db := l.svcCtx.DB.WithContext(ctx)
	records := make([]database.Comment, 0)
	err := db.Where("content_id = ? and root_id = ? and status = ? and created_at <= ?", contentId, 0, database.CommentStatusCommon, timestamp).
		Limit(limit).Order("created_at desc").Find(&records).Error
	if err != nil {
		logger.Error("get comment list by time from tidb:" + err.Error())
		return nil, err
	}
	logger.Info("get comment list by time from tidb")
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

func RecordsToResp(records []ListRecord) *commentRpc.CommentListResp {
	res := &commentRpc.CommentListResp{
		CommentId:   make([]int64, len(records)),
		UserId:      make([]int64, len(records)),
		ContentId:   make([]int64, len(records)),
		RootId:      make([]int64, len(records)),
		ParentId:    make([]int64, len(records)),
		CreatedAt:   make([]int64, len(records)),
		ShortText:   make([]string, len(records)),
		LongTextUri: make([]string, len(records)),
	}
	for i, v := range records {
		res.CommentId[i] = v.CommentId
		res.UserId[i] = v.UserId
		res.ContentId[i] = v.ContentId
		res.RootId[i] = v.RootId
		res.ParentId[i] = v.ParentId
		res.CreatedAt[i] = v.CreatedAt
		res.ShortText[i] = v.ShortText
		res.LongTextUri[i] = v.LongTextUri
	}
	return res
}

func RespGet(all *commentRpc.CommentListResp, limit int, timestamp int64) *commentRpc.CommentListResp {
	index := sort.Search(len(all.CommentId), func(i int) bool {
		return all.CreatedAt[i] <= timestamp
	})
	if index < 0 || len(all.CreatedAt) == 0 {
		return &commentRpc.CommentListResp{}
	}
	ma := index
	if index == len(all.CommentId) {
		ma--
	}
	mi := ma - limit + 1
	if mi < 0 {
		mi = 0
	}
	return &commentRpc.CommentListResp{
		CommentId:   all.CommentId[mi:ma],
		UserId:      all.UserId[mi:ma],
		ContentId:   all.ContentId[mi:ma],
		RootId:      all.RootId[mi:ma],
		ParentId:    all.ParentId[mi:ma],
		CreatedAt:   all.CreatedAt[mi:ma],
		ShortText:   all.ShortText[mi:ma],
		LongTextUri: all.LongTextUri[mi:ma],
	}

}
