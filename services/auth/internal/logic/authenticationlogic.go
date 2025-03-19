package logic

import (
	"context"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"time"

	"auth/internal/svc"
	"auth/proto/AuthRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthenticationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthenticationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthenticationLogic {
	return &AuthenticationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AuthenticationLogic) Authentication(in *AuthRpc.AuthenticationReq) (*AuthRpc.AuthenticationResp, error) {
	token, err := jwt.ParseWithClaims(in.Token, &svc.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return l.svcCtx.Secret, nil
	})
	if err == nil && !token.Valid {
		return &AuthRpc.AuthenticationResp{
			Pass: true,
		}, nil
	}

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	res, err := l.svcCtx.RDB.Get(timeout, in.SessionId).Result()
	cancel()
	if err != nil {
		return nil, err
	}

	s := &svc.Session{}
	err = json.Unmarshal([]byte(res), s)
	if err != nil {
		return nil, err
	}

	token = jwt.NewWithClaims(jwt.SigningMethodHS256, svc.JwtClaims{
		Userid: s.Userid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenStr, err := token.SignedString([]byte(l.svcCtx.Secret))
	if err != nil {
		return nil, err
	}

	return &AuthRpc.AuthenticationResp{
		Pass:  true,
		Token: tokenStr,
	}, nil

}
