package database

import (
	"time"
)

type VisibleContentInfo struct {
	Id      int64 `gorm:"PRIMARY_KEY"`
	Version int64 `gorm:"not null;default:1"`

	Userid       int64  `gorm:"not null;index:content,priority:10"`
	Title        string `gorm:"not null;index:content,priority:20"`
	PhotoUriList string `gorm:"index:content,priority:30"`

	ShortText    string
	LongTextUri  string
	VideoUriList string

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	DeletedAt time.Time `gorm:"autoDeleteTime"`
}

type InvisibleContentInfo struct {
	Id      int64 `gorm:"PRIMARY_KEY"`
	Version int64 `gorm:"not null;default:1"`

	Userid int64  `gorm:"not null"`
	Title  string `gorm:"not null"`

	PhotoUriList string
	ShortText    string
	LongTextUri  string
	VideoUriList string

	OldPhotoUriList string
	OldShortText    string
	OldLongTextUri  string
	OldVideoUriList string

	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
