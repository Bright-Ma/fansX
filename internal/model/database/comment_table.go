package database

import "time"

// Comment 若RootId==0 则该评论为根评论
type Comment struct {
	Id        int64 `gorm:"PRIMARY_KEY;"`
	UserId    int64 `gorm:"not null;"`
	ContentId int64 `gorm:"not null;index:time,priority:10;index:hot,priority:10"`
	RootId    int64 `gorm:"not null;index:time,priority:20;index:hot,priority:20"`
	Status    int   `gorm:"not null;index:time,priority:30;index:hot,priority:30"`
	CreatedAt int64 `gorm:"not null;index:time,priority:40;autoCreateTime"`
	Hot       int64 `gorm:"not null;index:hot,priority:40;default:100000"`

	ParentId  int64     `grom:"not null;"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime;"`

	ShortText   string `gorm:"not null;size:255"`
	LongTextUri string `gorm:"not null;size:255"`
}

type CommentCount struct {
	Id       int64 `gorm:"not null;PRIMARY_KEY"`
	Business int   `gorm:"not null;index:count,priority:10"`
	CountId  int64 `gorm:"not null;index:count,priority:20"`
	Status   int   `gorm:"not null;index:count,priority:30"`
	Count    int64 `gorm:"not null;index:count,priority:40"`
}

const (
	CommentStatusCommon = 1
	CommentStatusDelete = 2
)

const (
	CommentCountStatusCommon = 1
	CommentCountStatusDelete = 2
)

// comment幂等控制
// comment并发控制 kafka key有序
