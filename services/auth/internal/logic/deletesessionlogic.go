package logic

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"

	"auth/internal/svc"
	"auth/proto/AuthRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSessionLogic) DeleteSession(in *AuthRpc.DeleteSessionReq) (*AuthRpc.DeleteSessionResp, error) {
	rdb := l.svcCtx.RDB
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := rdb.Get(timeout, in.SessionId).Err()
	if err != nil && errors.Is(err, redis.Nil) {
		return &AuthRpc.DeleteSessionResp{}, nil
	} else if err != nil {
		return nil, err
	}

	err = rdb.Del(timeout, in.SessionId).Err()
	if err != nil {
		return nil, err
	}

	return &AuthRpc.DeleteSessionResp{}, nil
}
