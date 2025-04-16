package logic

import (
	"context"

	"fansX/services/like/internal/svc"
	"fansX/services/like/proto/likeRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikeNumsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLikeNumsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeNumsLogic {
	return &GetLikeNumsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLikeNumsLogic) GetLikeNums(in *likeRpc.GetLikeNumsReq) (*likeRpc.GetLikeNumsResp, error) {
	// todo: add your logic here and delete this line

	return &likeRpc.GetLikeNumsResp{}, nil
}
