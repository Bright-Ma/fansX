package svc

import (
	"encoding/json"
	"fansX/internal/middleware/lua"
	"fansX/internal/model/database"
	"fansX/pkg/hotkey-go/hotkey"
	leaf "fansX/pkg/leaf-go"
	syncx "fansX/pkg/sync"
	"fansX/services/comment/internal/config"
	"fansX/services/comment/proto/commentRpc"
	"github.com/IBM/sarama"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	etcd "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
	"log/slog"
	"strconv"
	"strings"
)

type ServiceContext struct {
	Config   config.Config
	Client   *redis.Client
	Executor *lua.Executor
	DB       *gorm.DB
	Producer sarama.SyncProducer
	Logger   *slog.Logger
	Creator  leaf.Core
	Cache    *hotkey.Core
	Sync     *syncx.Sync
	Group    *singleflight.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	hot := make(chan string, 1024*100)
	cache, err := hotkey.NewCore(hotkey.Config{
		Model:        hotkey.ModelCache,
		GroupName:    "",
		DelGroupName: "",
		CacheSize:    0,
		HotKeySize:   0,
		EtcdConfig:   etcd.Config{},
		HotChan:      hot,
	})
	if err != nil {
		panic(err.Error())
	}

	svc := &ServiceContext{
		Config: c,
		Cache:  cache,
	}
	go svc.addCache(hot)

	return svc
}

func (svc *ServiceContext) addCache(ch chan string) {
	db := svc.DB
	cache := svc.Cache
	for key := range ch {
		if strings.HasPrefix(key, "CommentListByHot:") {
			svc.addHot(key, db, cache)
		} else {
			svc.addTime(key, db, cache)
		}
	}
}

func (svc *ServiceContext) addHot(key string, db *gorm.DB, cache *hotkey.Core) {
	str, _ := strings.CutPrefix("CommentListByHot:", key)
	contentId, _ := strconv.ParseInt(str, 10, 64)
	records := make([]database.Comment, 0)
	err := db.Where("content_id = ? and root_id = ? and status= ?", contentId, 0, database.CommentStatusCommon).
		Limit(1000).Order("hot desc").Find(records).Error
	if err != nil {
		return
	}
	resp := commentRpc.CommentListResp{
		CommentId:   make([]int64, len(records)),
		UserId:      make([]int64, len(records)),
		ContentId:   make([]int64, len(records)),
		RootId:      make([]int64, len(records)),
		ParentId:    make([]int64, len(records)),
		CreatedAt:   make([]int64, len(records)),
		ShortText:   make([]string, len(records)),
		LongTextUri: make([]string, len(records)),
	}
	for i, v := range records {
		resp.CommentId[i] = v.Id
		resp.UserId[i] = v.UserId
		resp.RootId[i] = v.RootId
		resp.ParentId[i] = v.ParentId
		resp.CreatedAt[i] = v.CreatedAt
		resp.ShortText[i] = v.ShortText
		resp.LongTextUri[i] = v.LongTextUri
	}

	b, err := json.Marshal(&resp)
	if err != nil {
		return
	}
	cache.Set(key, b, 15)
	return

}

func (svc *ServiceContext) addTime(key string, db *gorm.DB, cache *hotkey.Core) {
	str, _ := strings.CutPrefix("CommentListByTime:", key)
	contentId, _ := strconv.ParseInt(str, 10, 64)
	records := make([]database.Comment, 0)
	err := db.Where("content_id = ? and root_id = ? and status= ?", contentId, 0, database.CommentStatusCommon).
		Limit(1000).Order("created_at desc").Find(records).Error
	if err != nil {
		return
	}
	resp := commentRpc.CommentListResp{
		CommentId:   make([]int64, len(records)),
		UserId:      make([]int64, len(records)),
		ContentId:   make([]int64, len(records)),
		RootId:      make([]int64, len(records)),
		ParentId:    make([]int64, len(records)),
		CreatedAt:   make([]int64, len(records)),
		ShortText:   make([]string, len(records)),
		LongTextUri: make([]string, len(records)),
	}
	for i, v := range records {
		resp.CommentId[i] = v.Id
		resp.UserId[i] = v.UserId
		resp.RootId[i] = v.RootId
		resp.ParentId[i] = v.ParentId
		resp.CreatedAt[i] = v.CreatedAt
		resp.ShortText[i] = v.ShortText
		resp.LongTextUri[i] = v.LongTextUri
	}

	b, err := json.Marshal(&resp)
	if err != nil {
		return
	}
	cache.Set(key, b, 15)
	return

}
