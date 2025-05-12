package logic

import (
	"context"
	"time"

	"fansX/services/user/internal/svc"
	"fansX/services/user/proto/UserRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LogoutLogic) Logout(in *UserRpc.LogoutReq) (*UserRpc.LogoutResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

}
