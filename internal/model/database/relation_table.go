package database

import "time"

type Following struct {
	Id          int64     `gorm:"PRIMARY_KEY"`
	FollowerId  int64     `gorm:"not null;index:following,priority:10"`
	Type        int       `gorm:"not null;index:following,priority:20"`
	FollowingId int64     `gorm:"not null;index:following,priority:30"`
	UpdatedAt   int64     `gorm:"not null;index:following,priority:40;autoUpdateTime:nano"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type FollowingNums struct {
	UserId int64 `gorm:"PRIMARY_KEY"`
	Nums   int64 `gorm:"not null;default:0"`
}

type Follower struct {
	Id          int64     `gorm:"PRIMARY_KEY"`
	FollowingId int64     `gorm:"not null;index:follower,priority:10"`
	Type        int       `gorm:"not null;index:follower,priority:20"`
	FollowerId  int64     `gorm:"not null;index:follower,priority:30"`
	UpdatedAt   int64     `gorm:"not null;index:follower,priority:40;autoUpdateTime:nano"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

type FollowerNums struct {
	UserId int64 `gorm:"PRIMARY_KEY"`
	Nums   int64 `gorm:"not null;default:0"`
}

var (
	Followed   = 1
	UnFollowed = 0
)
