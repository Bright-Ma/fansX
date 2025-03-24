package model

import "time"

type Following struct {
	Id          int64     `gorm:"PRIMARY_KEY"`
	FollowerId  int64     `gorm:"index:following,priority:10"`
	FollowingId int64     `gorm:"index:following,priority:20"`
	Type        int       `gorm:"index:following,priority:30"`
	UpdatedAt   time.Time `gorm:"index:following,priority:40"`
}

type FollowingNums struct {
	UserId int64 `gorm:"PRIMARY_KEY"`
	Nums   int64 `gorm:"not null;default:0"`
}

type Follower struct {
	Id          int64     `gorm:"PRIMARY_KEY"`
	FollowingId int64     `gorm:"index:follower,priority:10"`
	FollowerId  int64     `gorm:"index:follower,priority:20"`
	Type        int       `gorm:"index:follower,priority:30"`
	UpdateAt    time.Time `gorm:"index:follower,priority:40"`
}

type FollowerNums struct {
	UserId int64 `gorm:"PRIMARY_KEY"`
	Nums   int64 `gorm:"not null;default:0"`
}

var (
	Followed   = true
	UnFollowed = false
)
