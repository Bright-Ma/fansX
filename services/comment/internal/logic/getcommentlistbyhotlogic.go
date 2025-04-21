package logic

import (
	"context"
	"fansX/internal/util"
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

func (l *GetCommentListByHotLogic) GetCommentListByHot(in *commentRpc.GetCommentListByHotReq) (*commentRpc.CommentListResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	logger.Info("user get comment list by hot", "contentId", in.ContentId, "limit", in.Limit, "offset", in.Offset)
	cache := l.svcCtx.Cache

	key := "CommentListByHot:" + strconv.FormatInt(in.ContentId, 10)
	v, ok := cache.Get(key)
	if ok {

	}
}
