package database

import "time"

// Like
// index user-> business status userId updatedAt likeId
// index like-> business status likeId updatedAt userId
type Like struct {
	Id int64 `gorm:"PRIMARY_KEY"`

	Business  int   `gorm:"not null;index:user,priority:10;index:like,priority:10;"`
	Status    int   `gorm:"not null;index:user,priority:20;index:like,priority:20;"`
	UserId    int64 `gorm:"not null;index:user,priority:30;index:like,priority:50;"`
	UpdatedAt int64 `gorm:"not null;index:user,priority:40;index:like,priority:40;"`
	LikeId    int64 `gorm:"not null;index:user,priority:50;index:like,priority:30;"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type LikeCount struct {
	Id       int64 `gorm:"PRIMARY_KEY"`
	Business int   `gorm:"not null;index:like,priority:10"`
	LikeId   int64 `gorm:"not null;index:like,priority:20"`
	Count    int64 `gorm:"not null;index:like,priority:30;default:0"`
}

const (
	BusinessContent = 1
	BusinessComment = 2
)

const (
	LikeStatusLike   = 1
	LikeStatusUnlike = 0
)

/*
	以下弃用
*/

type TimeWindow struct {
	Id       int64 `gorm:"PRIMARY_KEY"`
	Business int64 `gorm:"not null;index:window,priority:10"`
	Window   int64 `gorm:"index:window,priority:20"`
	Status   int64 `gorm:"not null;index:window,priority:30"`

	Body []byte `gorm:"size:204800"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type WindowLatest struct {
	Id         int64 `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Business   int64 `gorm:"not null"`
	NextWindow int64 `gorm:"not null"`
}

type WindowBody struct {
	//
	Like map[int64][]int64 `json:"like"`
}
