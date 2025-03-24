package logic

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"time"

	"auth/internal/svc"
	"auth/proto/AuthRpc"

	"encoding/json"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshSessionLogic {
	return &RefreshSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshSessionLogic) RefreshSession(in *AuthRpc.RefreshSessionReq) (*AuthRpc.RefreshSessionResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	ok, err := l.svcCtx.RDB.Expire(timeout, in.SessionId, time.Hour*24*7).Result()
	cancel()
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("redis expire fail,key not exits")
	}

	timeout, cancel = context.WithTimeout(context.Background(), time.Second)
	res, err := l.svcCtx.RDB.Get(timeout, in.SessionId).Result()
	cancel()
	if err != nil {
		return nil, err
	}

	s := &svc.Session{}
	err = json.Unmarshal([]byte(res), &s)
	if err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, svc.JwtClaims{
		Userid: s.Userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenStr, err := token.SignedString(l.svcCtx.Secret)
	if err != nil {
		return &AuthRpc.RefreshSessionResp{}, nil
	}

	return &AuthRpc.RefreshSessionResp{
		Ok:    true,
		Token: tokenStr,
	}, nil
}
