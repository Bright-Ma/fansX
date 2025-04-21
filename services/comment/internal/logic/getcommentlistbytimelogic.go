package logic

import (
	"context"

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

func (l *GetCommentListByTimeLogic) GetCommentListByTime(in *commentRpc.GetCommentListByTimeReq) (*commentRpc.CommentListResp, error) {
	// todo: add your logic here and delete this line

	return &commentRpc.CommentListResp{}, nil
}
