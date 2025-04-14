package main

import (
	"context"
	luaString "fansX/common/lua/script/string"
	"fansX/internal/model/database"
	"fansX/internal/model/mq"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func (h *Handler) UpdateNums(tx *gorm.DB, data *database.Follower) error {
	if data.Type == database.Followed {
		return tx.Take(&database.FollowerNums{}, data.FollowingId).Update("nums", gorm.Expr("nums + 1")).Error
	} else {
		return tx.Take(&database.FollowerNums{}, data.FollowingId).Update("nums", gorm.Expr("nums - 1")).Error
	}
}

func (h *Handler) UpdateRedis(data *database.Follower) {
	e := h.executor
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	key := "Follower:" + strconv.FormatInt(data.FollowingId, 10)
	if data.Type == database.Followed {
		e.Execute(timeout, luaFollowing.GetAdd(), []string{key}, strconv.FormatInt(data.UpdatedAt, 10), strconv.FormatInt(data.FollowerId, 10))
		e.Execute(timeout, luaString.GetIncrBy(), []string{"FollowerNums:" + strconv.FormatInt(data.FollowingId, 10)}, 1)
	} else {
		h.client.ZRem(timeout, key, strconv.FormatInt(data.FollowerId, 10))
		e.Execute(timeout, luaString.GetIncrBy(), []string{"FollowerNums:" + strconv.FormatInt(data.FollowingId, 10)}, -1)
	}
}

func Trans(f *mq.Following) *database.Follower {
	t, _ := strconv.Atoi(f.Type)
	id, _ := strconv.ParseInt(f.Id, 10, 64)
	u, _ := strconv.ParseInt(f.UpdatedAt, 10, 64)
	followerId, _ := strconv.ParseInt(f.FollowerId, 10, 64)
	followingId, _ := strconv.ParseInt(f.FollowingId, 10, 64)

	return &database.Follower{
		Id:          id,
		FollowerId:  followerId,
		Type:        t,
		FollowingId: followingId,
		UpdatedAt:   u,
	}
}
