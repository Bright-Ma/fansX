package logic

import (
	"context"

	"bilibili/services/feed/internal/svc"
	"bilibili/services/feed/proto/feedRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PullLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPullLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PullLogic {
	return &PullLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PullLogic) Pull(in *feedRpc.PullReq) (*feedRpc.PullResp, error) {
	// todo: add your logic here and delete this line

	return &feedRpc.PullResp{}, nil
}
