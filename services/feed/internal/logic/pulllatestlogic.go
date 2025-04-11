package logic

import (
	"context"

	"bilibili/services/feed/internal/svc"
	"bilibili/services/feed/proto/feedRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PullLatestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPullLatestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PullLatestLogic {
	return &PullLatestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PullLatestLogic) PullLatest(in *feedRpc.PullLatestReq) (*feedRpc.PullResp, error) {
	rdb := l.svcCtx.RClient

}
