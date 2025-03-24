package logic

import (
	"context"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowingLogic {
	return &ListFollowingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFollowingLogic) ListFollowing(in *relationRpc.ListFollowingReq) (*relationRpc.ListFollowingResp, error) {
	// todo: add your logic here and delete this line

	return &relationRpc.ListFollowingResp{}, nil
}
