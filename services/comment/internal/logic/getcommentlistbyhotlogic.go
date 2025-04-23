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

type GetCommentListByHotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentListByHotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentListByHotLogic {
	return &GetCommentListByHotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

type ListHotRecord struct {
	CommentId   int64  `json:"comment_id"`
	UserId      int64  `json:"user_id"`
	ContentId   int64  `json:"content_id"`
	RootId      int64  `json:"root_id"`
	ParentId    int64  `json:"parent_id"`
	CreatedAt   int64  `json:"created_at"`
	ShortText   string `json:"short_text"`
	LongTextUri string `json:"long_text_uri"`
	hot         int64
}

func (l *GetCommentListByHotLogic) GetCommentListByHot(in *commentRpc.GetCommentListByHotReq) (*commentRpc.CommentListResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user get comment list by hot", "contentId", in.ContentId, "limit", in.Limit, "offset", in.Offset)
	cache := l.svcCtx.Cache

	key := "CommentListByHot:" + strconv.FormatInt(in.ContentId, 10)
	v, ok := cache.Get(key)
	if ok {
		logger.Info("get comment list by hot from local cache")
		value := &commentRpc.CommentListResp{}
		err := json.Unmarshal(v, &value)
		if err != nil {
			panic(err.Error())
		}
		return l.HotGet(value, int(in.Limit), int(in.Offset)), nil
	}

	records, status := l.GetFromRedis(timeout, key, int(in.Limit), int(in.Offset), logger)
	if status == StatusFind {
		return l.HotToResp(records), nil
	} else if status == StatusNeedRebuild {
		go func() {
			mutex := l.svcCtx.Sync.NewMutex(key+":mutex", syncx.WithTTL(time.Second), syncx.WithUtil(time.Second*5))
			if mutex.TryLock() == nil {
				records, err := l.GetFromTiDB(timeout, in, 1000, 0, logger)
				if err == nil {
					l.BuildRedis(key, records)
				}
				_ = mutex.Unlock()
			}
		}()
		return l.HotToResp(records), nil
	} else if status == StatusError {
		records, err := l.GetFromTiDB(timeout, in, int(in.Limit), int(in.Offset), logger)
		if err != nil {
			return nil, err
		}
		return l.HotToResp(records), nil
	}
	records, err := l.GetFromTiDB(timeout, in, 1000, 0, logger)
	if err != nil {
		return nil, err
	}
	go l.BuildRedis(key, records)
	return l.HotGet(l.HotToResp(records), int(in.Limit), int(in.Offset)), nil
}

func (l *GetCommentListByHotLogic) BuildRedis(key string, records []ListHotRecord) {
	data := make([]interface{}, len(records)*2+2)
	for i, v := range records {
		b, err := json.Marshal(records[i])
		if err != nil {
			panic(err.Error())
		}
		data[i*2+1] = string(b)
		data[i*2] = v.hot
	}
	data[len(data)-2] = -1
	data[len(data)-1] = 5
	err := l.svcCtx.Executor.Execute(context.Background(), commentservicescript.Build, []string{key, strconv.Itoa(60)}, data).Err()
	if err != nil {
		slog.Error("get comment list by hot build redis:" + err.Error())
	}
}

func (l *GetCommentListByHotLogic) GetFromRedis(ctx context.Context, key string, limit int, offset int, logger *slog.Logger) ([]ListHotRecord, int) {
	executor := l.svcCtx.Executor
	resp, err := executor.Execute(ctx, commentservicescript.GetByHot, []string{key}, limit, offset).Result()
	if err != nil {
		logger.Error("execute lua to get hot list from redis:" + err.Error())
		return nil, StatusError
	}

	listInter := resp.([]interface{})
	status, _ := strconv.ParseInt(listInter[len(listInter)-1].(string), 10, 64)
	if status == StatusNotFind {
		logger.Debug("hot comment list not exists from redis")
		return nil, int(status)
	} else if status == StatusNeedRebuild {
		logger.Info("get hot comment list from redis but need to rebuild")
	} else {
		logger.Info("get hot comment list from redis")
	}
	res := make([]ListHotRecord, len(listInter)-1)

	for i := 0; i < len(listInter)-1; i++ {
		err = json.Unmarshal([]byte(listInter[i].(string)), &res[i])
		if err != nil {
			panic(err.Error())
		}
	}

	return res, int(status)
}

func (l *GetCommentListByHotLogic) GetFromTiDB(ctx context.Context, in *commentRpc.GetCommentListByHotReq, limit int, offset int, logger *slog.Logger) ([]ListHotRecord, error) {
	db := l.svcCtx.DB.WithContext(ctx)
	records := make([]database.Comment, 0)
	err := db.Where("content_id = ? and root_id = ? and status = ?", in.ContentId, 0, database.CommentStatusCommon).
		Limit(limit).Offset(offset).Order("hot desc").Find(&records).Error
	if err != nil {
		logger.Error("get comment list by hot from tidb:" + err.Error())
		return nil, err
	}
	logger.Info("get comment list by hot from tidb")
	res := make([]ListHotRecord, len(records))
	for i, v := range records {
		res[i].hot = v.Hot
		res[i].CommentId = v.Id
		res[i].ContentId = v.ContentId
		res[i].ParentId = v.ParentId
		res[i].UserId = v.UserId
		res[i].LongTextUri = v.LongTextUri
		res[i].ShortText = v.ShortText
		res[i].CreatedAt = v.CreatedAt
		res[i].RootId = v.RootId
	}

	return res, nil
}

func (l *GetCommentListByHotLogic) HotToResp(list []ListHotRecord) *commentRpc.CommentListResp {
	res := &commentRpc.CommentListResp{
		CommentId:   make([]int64, len(list)),
		UserId:      make([]int64, len(list)),
		ContentId:   make([]int64, len(list)),
		RootId:      make([]int64, len(list)),
		ParentId:    make([]int64, len(list)),
		CreatedAt:   make([]int64, len(list)),
		ShortText:   make([]string, len(list)),
		LongTextUri: make([]string, len(list)),
	}
	for i, v := range list {
		res.CommentId[i] = v.CommentId
		res.UserId[i] = v.CommentId
		res.ContentId[i] = v.ContentId
		res.RootId[i] = v.RootId
		res.ParentId[i] = v.ParentId
		res.CreatedAt[i] = v.CreatedAt
		res.ShortText[i] = v.ShortText
		res.LongTextUri[i] = v.LongTextUri
	}
	return res
}

func (l *GetCommentListByHotLogic) HotGet(all *commentRpc.CommentListResp, limit int, offset int) *commentRpc.CommentListResp {
	if len(all.CreatedAt) == 0 || offset >= len(all.CreatedAt) {
		return &commentRpc.CommentListResp{}
	}
	var ma = min(len(all.CreatedAt)-1, limit+offset-1)
	res := &commentRpc.CommentListResp{
		CommentId:   all.CommentId[offset:ma],
		UserId:      all.UserId[offset:ma],
		ContentId:   all.ContentId[offset:ma],
		RootId:      all.RootId[offset:ma],
		ParentId:    all.ParentId[offset:ma],
		CreatedAt:   all.CreatedAt[offset:ma],
		ShortText:   all.ShortText[offset:ma],
		LongTextUri: all.LongTextUri[offset:ma],
	}

	return res
}
