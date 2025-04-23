package logic

import (
	"context"
	"encoding/json"
	"fansX/internal/model/database"
	"fansX/internal/script/commentservicescript"
	"fansX/internal/util"
	syncx "fansX/pkg/sync"
	"log/slog"
	"slices"
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

type ListTimeRecord struct {
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
	}

	records, status := l.GetFromRedis(timeout, key, int(in.Limit), in.TimeStamp, logger)
	if status == StatusFind {
		return l.TimeToResp(records), nil
	} else if status == StatusNeedRebuild {
		go func() {
			records, err := l.GetFromTiDB(timeout, in, 1000, 0, logger)
			if err == nil {
				l.BuildRedis(key, records)
			}
		}()
		return l.TimeToResp(records), nil
	} else if status == StatusError {
		records, err := l.GetFromTiDB(timeout, in, int(in.Limit), in.TimeStamp, logger)
		if err != nil {
			return nil, err
		}
		return l.TimeToResp(records), nil
	}
	records, err := l.GetFromTiDB(timeout, in, int(in.Limit), time.Now().Add(time.Hour).Unix(), logger)
	if err != nil {
		return nil, err
	}
	go l.BuildRedis(key, records)
	return l.TimeGet(l.TimeToResp(records), int(in.Limit), in.TimeStamp), nil

}

func (l *GetCommentListByTimeLogic) BuildRedis(key string, records []ListTimeRecord) {
	mutex := l.svcCtx.Sync.NewMutex(key, syncx.WithUtil(time.Second*5), syncx.WithTTL(time.Second))
	err := mutex.TryLock()
	if err != nil {
		return
	}
	data := make([]interface{}, len(records)*2+2)
	for i, v := range records {
		b, err := json.Marshal(records[i])
		if err != nil {
			panic(err.Error())
		}
		data[i*2+1] = string(b)
		data[i*2] = v.CreatedAt
	}
	err = l.svcCtx.Executor.Execute(context.Background()).Err()
	if err != nil {
		slog.Error("build comment list by time redis:" + err.Error())
	}
	_ = mutex.Unlock()
}

func (l *GetCommentListByTimeLogic) GetFromRedis(ctx context.Context, key string, limit int, timestamp int64, logger *slog.Logger) ([]ListTimeRecord, int) {
	executor := l.svcCtx.Executor
	resp, err := executor.Execute(ctx, commentservicescript.GetByTime, []string{key}, limit, timestamp).Result()
	if err != nil {
		logger.Error("execute lua to get time list from redis:" + err.Error())
		return nil, StatusError
	}

	listInter := resp.([]interface{})
	status, _ := strconv.ParseInt(listInter[len(listInter)-1].(string), 10, 64)
	if status == StatusNotFind {
		logger.Debug("time comment list not exists from redis")
		return nil, int(status)
	} else if status == StatusNeedRebuild {
		logger.Info("get time comment list from redis but need to rebuild")
	} else {
		logger.Info("get time comment list from redis")
	}
	res := make([]ListTimeRecord, len(listInter)-1)
	for i := 0; i < len(listInter)-1; i++ {
		err = json.Unmarshal([]byte(listInter[i].(string)), &res[i])
		if err != nil {
			panic(err.Error())
		}
	}
	return res, int(status)
}

func (l *GetCommentListByTimeLogic) GetFromTiDB(ctx context.Context, in *commentRpc.GetCommentListByTimeReq, limit int, timestamp int64, logger *slog.Logger) ([]ListTimeRecord, error) {
	db := l.svcCtx.DB.WithContext(ctx)
	records := make([]database.Comment, 0)
	err := db.Where("content_id = ? and root_id = ? and status = ? and created_at <= ?", in.ContentId, 0, database.CommentStatusCommon, timestamp).
		Limit(limit).Order("created_at desc").Find(&records).Error
	if err != nil {
		logger.Error("get comment list by time from tidb:" + err.Error())
		return nil, err
	}
	logger.Info("get comment list by time from tidb")
	res := make([]ListTimeRecord, len(records))
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

func (l *GetCommentListByTimeLogic) TimeToResp(records []ListTimeRecord) *commentRpc.CommentListResp {
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

func (l *GetCommentListByTimeLogic) TimeGet(all *commentRpc.CommentListResp, limit int, timestamp int64) *commentRpc.CommentListResp {
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
