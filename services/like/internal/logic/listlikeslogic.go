package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &likeRpc.ListLikesResp{}, nil
}
