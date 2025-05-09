package logic

import (
	"context"
	"fansX/internal/model/database"
	"gorm.io/gorm/logger"
	"hash/fnv"
	"time"

	"fansX/services/user/internal/svc"
	"fansX/services/user/proto/UserRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *UserRpc.LoginReq) (*UserRpc.LoginResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
}
