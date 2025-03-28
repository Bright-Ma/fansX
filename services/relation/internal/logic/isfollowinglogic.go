package logic

import (
	luaZset "bilibili/common/lua/script/zset"
	"bilibili/common/util"
	"bilibili/internal/model/database"
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"time"

	"bilibili/services/relation/internal/svc"
	"bilibili/services/relation/proto/relationRpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsFollowingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIsFollowingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsFollowingLogic {
	return &IsFollowingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IsFollowingLogic) IsFollowing(in *relationRpc.IsFollowingReq) (*relationRpc.IsFollowingResp, error) {
	db := l.svcCtx.DB
	logger := util.SetTrace(l.ctx, l.svcCtx.Logger)
	s := l.svcCtx.Single
	executor := l.svcCtx.Executor

	logger.Info("judge is following", "userId", in.UserId, "followingId", in.FollowId)
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "Following:" + strconv.FormatInt(in.UserId, 10)
	res, err := executor.Execute(timeout, luaZset.GetGetField(), []string{key}, strconv.FormatInt(in.FollowId, 10)).Result()
	if err != nil {
		logger.Error("exec script GetField:" + err.Error())
		return nil, err
	}

	if res.(string) == luaZset.FieldNotExists {
		logger.Debug("field not exists")
		return &relationRpc.IsFollowingResp{Is: false}, nil
	} else if res.(string) != luaZset.TableNotExists {
		logger.Debug("field exists")
		return &relationRpc.IsFollowingResp{Is: true}, nil
	}

	logger.Debug("table not exists")

	followed, err := s.Do("IsFollowing:"+strconv.FormatInt(in.UserId, 10), func() (interface{}, error) {
		record := make([]database.Following, 0)
		timeout, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = db.WithContext(timeout).Select("following_id", "updated_at").
			Where("follower_id =  ? and type = ?", in.UserId, database.Followed).
			Find(&record).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		followed := make(map[int64]bool)
		for _, v := range record {
			followed[v.FollowingId] = true
		}

		go func() {
			kvs := make([]interface{}, len(record)*2)
			for i, v := range record {
				kvs[i*2] = strconv.FormatInt(v.UpdatedAt, 10)
				kvs[i*2+1] = strconv.FormatInt(v.FollowingId, 10)
			}
			timeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			err := executor.Execute(timeout, luaZset.GetCreate(), []string{key, "false"}, kvs).Err()
			if err != nil {
				logger.Warn("execute zset_create:" + err.Error())
			}
		}()
		return followed, err
	})

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("search table-followings:" + err.Error())
		return nil, err
	}
	_, ok := followed.(map[int64]bool)[in.FollowId]

	if !ok {
		logger.Debug("search table-followings not exists following relation")
		return &relationRpc.IsFollowingResp{Is: false}, nil
	} else {
		logger.Debug("search table-followings exists following relation")
		return &relationRpc.IsFollowingResp{Is: true}, nil
	}
}
