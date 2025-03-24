package logic

import (
	"context"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFollowerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFollowerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFollowerLogic {
	return &ListFollowerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFollowerLogic) ListFollower(in *relationRpc.ListFollowerReq) (*relationRpc.ListFollowerResp, error) {
	// todo: add your logic here and delete this line

	return &relationRpc.ListFollowerResp{}, nil
}
