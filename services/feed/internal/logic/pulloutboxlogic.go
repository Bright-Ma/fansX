package logic

import (
	"context"

	"bilibili/services/feed/internal/svc"
	"bilibili/services/feed/proto/feedRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PullOutBoxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPullOutBoxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PullOutBoxLogic {
	return &PullOutBoxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PullOutBoxLogic) PullOutBox(in *feedRpc.PullOutBoxReq) (*feedRpc.PullOutBoxResp, error) {
	// todo: add your logic here and delete this line

	return &feedRpc.PullOutBoxResp{}, nil
}
