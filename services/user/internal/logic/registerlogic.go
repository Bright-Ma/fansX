package logic

import (
	"context"
	"errors"
	"fansX/internal/model/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"hash/fnv"
	"time"

	"fansX/services/user/internal/svc"
	"fansX/services/user/proto/UserRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *UserRpc.RegisterReq) (*UserRpc.RegisterResp, error) {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	creator := l.svcCtx.Creator
	logger := l.svcCtx.Logger
	tx := l.svcCtx.DB.WithContext(timeout).Begin()
	logger.Info("user register", "phone", in.Phone, "name", in.UserName)
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("phone = ?", in.Phone).
		Take(&database.User{}).Error
	if err == nil {
		tx.Commit()
		logger.Info("register failed because duplicate phone")
		return nil, errors.New("duplicate phone")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Commit()
		logger.Info("register failed search phone from tidb:" + err.Error())
		return nil, err
	}
	id, err := creator.GetIdWithContext(timeout)
	if err != nil {
		logger.Error("register failed get id with context:" + err.Error())
		return nil, err
	}

	h := fnv.New128a()
	h.Write(append([]byte(in.PassWord), salt...))
	b := make([]byte, 128)
	sum := h.Sum(b)

	record := database.User{
		Id:       id,
		Phone:    in.Phone,
		Name:     in.UserName,
		Password: string(sum),
	}
	err = tx.Create(&record).Error
	if err != nil {
		tx.Rollback()
		logger.Error("register failed insert record to tidb:" + err.Error())
		return nil, err
	}
	tx.Commit()
	logger.Info("register success")
	return &UserRpc.RegisterResp{}, nil
}

var salt = []byte("salt")
