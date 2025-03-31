package logic

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	cshash "puhser/consistenthash"
	"puhser/mq"
	"puhser/proto/push"
	"puhser/route"
	"strconv"

	"puhser/getway/in/internal/svc"
	"puhser/getway/in/proto/InnerGetWay"

	"github.com/zeromicro/go-zero/core/logx"
)

type PushMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPushMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PushMessageLogic {
	return &PushMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PushMessageLogic) PushMessage(in *InnerGetWay.PushMessageReq) (*InnerGetWay.PushMessageResp, error) {
	ctx := l.svcCtx
	if in.Global {
		return PushMessageGlobal(ctx, in)
	}

	if ctx.Config.Model == 2 {
		return ConsistentHashPush(ctx, in)
	}

	if len(in.UserId) == 1 {
		return PushMessageOnce(ctx, in)
	}

	return PushMessage(ctx, in)
}

func ConsistentHashPush(ctx *svc.ServiceContext, in *InnerGetWay.PushMessageReq) (*InnerGetWay.PushMessageResp, error) {

	UserIdStr := make([]string, len(in.UserId))
	for index, id := range in.UserId {
		UserIdStr[index] = strconv.FormatInt(id, 10)
	}

	mp := make(map[*grpc.ClientConn][]int64)
	addr := cshash.Get(UserIdStr)
	for i, v := range addr {
		conn, ok := ctx.Services.Load(v)
		if !ok {
			mp[nil] = append(mp[nil], in.UserId[i])
			continue
		}
		c, _ := conn.(*grpc.ClientConn)
		mp[c] = append(mp[c], in.UserId[i])
	}

	for k, v := range mp {
		if k == nil {
			continue
		}
		req := &push.PushMessageReq{
			UserId:     v,
			PayLoad:    in.PayLoad,
			EncodeType: in.EncodeType,
		}
		client := push.NewPushMessageServiceClient(k)
		_, _ = client.PushMessage(context.Background(), req)
	}

	return &InnerGetWay.PushMessageResp{}, nil

}

func PushMessageGlobal(ctx *svc.ServiceContext, in *InnerGetWay.PushMessageReq) (*InnerGetWay.PushMessageResp, error) {
	req := mq.Request{
		Msg: route.Message{
			UUId:       uuid.New().String(),
			PayLoad:    in.PayLoad,
			EncodeType: in.EncodeType,
		},
	}

	for i := 0; i < 100; i++ {
		req.BucketId = int64(i)

		msg, err := json.Marshal(req)
		if err != nil {
			continue
		}

		_ = ctx.Producer.Publish(ctx.Config.NSQ.Topic, msg)
	}

	return &InnerGetWay.PushMessageResp{}, nil

}

func PushMessageOnce(ctx *svc.ServiceContext, in *InnerGetWay.PushMessageReq) (*InnerGetWay.PushMessageResp, error) {
	req := push.PushMessageReq{
		UserId:     in.UserId,
		PayLoad:    in.PayLoad,
		EncodeType: in.EncodeType,
	}

	v, ok := ctx.Services.Load(ctx.RDB.Get(context.Background(), strconv.FormatInt(in.UserId[0], 10)).String())
	if !ok {
		return &InnerGetWay.PushMessageResp{}, errors.New("disConnect")
	}

	client, ok := v.(push.PushMessageServiceClient)
	if !ok {
		return &InnerGetWay.PushMessageResp{}, errors.New("err interface")
	}

	_, err := client.PushMessage(context.Background(), &req)
	if err != nil {
		return &InnerGetWay.PushMessageResp{}, err
	}

	return &InnerGetWay.PushMessageResp{}, nil
}

func PushMessage(ctx *svc.ServiceContext, in *InnerGetWay.PushMessageReq) (*InnerGetWay.PushMessageResp, error) {
	mp := make(map[interface{}][]int64)
	UserId := make([]string, len(in.UserId))

	for i, v := range in.UserId {
		UserId[i] = strconv.FormatInt(v, 10)
	}

	add := 50
	for i := 0; i < len(UserId); {

		var ids []string
		if i+add < len(UserId) {
			ids = UserId[i : i+add]
		} else {
			ids = UserId[i:]
		}

		res, err := ctx.RDB.MGet(context.Background(), ids...).Result()
		if err != nil {
			continue
		}

		for j, v := range res {
			client, ok := ctx.Services.Load(v)
			if !ok {
				continue
			}
			mp[client] = append(mp[client], in.UserId[i+j])
		}

		i = min(i+add+1, len(UserId))
	}

	for k, v := range mp {
		_, _ = k.(push.PushMessageServiceClient).PushMessage(context.Background(), &push.PushMessageReq{
			UserId:     v,
			PayLoad:    in.PayLoad,
			EncodeType: in.EncodeType,
		})
	}

	return &InnerGetWay.PushMessageResp{}, nil

}
