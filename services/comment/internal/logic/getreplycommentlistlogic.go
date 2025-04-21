package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &commentRpc.CommentListResp{}, nil
}
