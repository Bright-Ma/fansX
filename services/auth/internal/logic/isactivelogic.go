package logic

import (
	"context"

	"auth/internal/svc"
	"auth/proto/AuthRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsActiveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsActiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsActiveLogic {
	return &IsActiveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsActiveLogic) IsActive(in *AuthRpc.IsActiveReq) (*AuthRpc.IsActiveResp, error) {
	// todo: add your logic here and delete this line

	return &AuthRpc.IsActiveResp{}, nil
}
