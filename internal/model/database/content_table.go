package database

import (
	"time"
)

type VisibleContentInfo struct {
	Id      int64 `gorm:"PRIMARY_KEY"`
	Version int64 `gorm:"not null;default:1"`

	Userid       int64  `gorm:"not null;index"`
	Status       int    `gorm:"not null"`
	Title        string `gorm:"size:255"`
	PhotoUriList string `gorm:"size:8192"`

	ShortText    string `gorm:"size:255"`
	LongTextUri  string `gorm:"size:255"`
	VideoUriList string `gorm:"size:8192"`

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedAt int64     `gorm:"autoCreateTime:milli"`
}

type InvisibleContentInfo struct {
	Id      int64 `gorm:"PRIMARY_KEY" json:"id"`
	Version int64 `gorm:"not null;default:1" json:"version"`

	Status    int    `gorm:"not null" json:"status"`
	OldStatus int    `gorm:"not null" json:"-"`
	Desc      string `gorm:"size:255" json:"-"`

	Userid int64  `gorm:"not null" json:"user_id"`
	Title  string `gorm:"not null;size:255" json:"title"`

	PhotoUriList string `gorm:"size:8192" json:"photo_uri_list"`
	ShortText    string `gorm:"size:255" json:"short_text"`
	LongTextUri  string `gorm:"size:255" json:"long_text_uri"`
	VideoUriList string `gorm:"size:8192" json:"video_uri_list"`

	OldPhotoUriList string `gorm:"size:8192" json:"-"`
	OldShortText    string `gorm:"size:255" json:"-"`
	OldLongTextUri  string `gorm:"size:255" json:"-"`
	OldVideoUriList string `gorm:"size:8192" json:"-"`

	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
}

var (
	ContentStatusCheck   = 1
	ContentStatusPass    = 2
	ContentStatusNotPass = 3
	ContentStatusDelete  = 4
)
