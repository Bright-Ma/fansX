package logic

import (
	"context"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"

	"auth/internal/svc"
	"auth/proto/AuthRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateVoucherLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateVoucherLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateVoucherLogic {
	return &CreateVoucherLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateVoucherLogic) CreateVoucher(in *AuthRpc.CreateVoucherReq) (*AuthRpc.CreateVoucherResp, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, svc.JwtClaims{
		Userid: in.UserId,
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

	s := svc.Session{Userid: in.UserId}
	sessionId := uuid.New().String()
	js, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	err = l.svcCtx.RDB.Set(timeout, sessionId, string(js), time.Hour*24*7).Err()
	cancel()
	if err != nil {
		return nil, err
	}

	return &AuthRpc.CreateVoucherResp{
		Ok:        true,
		SessionId: sessionId,
		Token:     tokenStr,
	}, nil
}
