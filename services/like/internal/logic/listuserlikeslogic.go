package logic

import (
	"context"

	"fansX/services/like/internal/svc"
	"fansX/services/like/proto/likeRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserLikesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserLikesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserLikesLogic {
	return &ListUserLikesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListUserLikesLogic) ListUserLikes(in *likeRpc.ListUserLikesReq) (*likeRpc.ListUserLikesResp, error) {
	// todo: add your logic here and delete this line

	return &likeRpc.ListUserLikesResp{}, nil
}
