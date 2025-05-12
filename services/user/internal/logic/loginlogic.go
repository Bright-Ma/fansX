package logic

import (
	"context"
	"errors"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"fansX/services/auth/proto/AuthRpc"
	"github.com/IBM/sarama"
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

	db := l.svcCtx.DB.WithContext(timeout)
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	aClient := l.svcCtx.AuthClient
	producer := l.svcCtx.Producer
	logger.Info("user login", "phone", in.Phone)

	h := fnv.New128a()
	h.Write(append([]byte(in.PassWord), salt...))
	b := make([]byte, 128)
	h.Sum(b)

	u := &database.User{}
	err := db.Where("phone = ?", in.Phone).Take(u).Error
	if err != nil {
		logger.Error("user login:" + err.Error())
		return nil, err
	}
	if u.Password != string(b) {
		err = errors.New("password error")
		logger.Error("user login:"+err.Error(), "password", string(b))
		return nil, err
	}
	resp, err := aClient.CreateVoucher(l.ctx, &AuthRpc.CreateVoucherReq{UserId: u.Id})
	if err != nil {
		logger.Error("user login create voucher:" + err.Error())
		return nil, err
	}
	logger.Info("user login success")
	producer.SendMessage(&sarama.ProducerMessage{
		Topic:     "",
		Key:       nil,
		Value:     nil,
		Headers:   nil,
		Metadata:  nil,
		Offset:    0,
		Partition: 0,
		Timestamp: time.Time{},
	})
	return &UserRpc.LoginResp{
		SessionId: resp.SessionId,
		Token:     resp.Token,
	}, nil

}
