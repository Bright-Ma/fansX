package segment

import (
	"gorm.io/gorm"
	"sync"
	"time"
)

type Creator struct {
	id  int64
	mu  sync.Mutex
	db  *gorm.DB
	ch  chan int
	old *buffer
	new *buffer
}

type buffer struct {
	nextId   int64
	maxId    int64
	preIndex int64
}

type IdTable struct {
	ID       int64     `gorm:"primary_key"`
	Tag      string    `gorm:"unique_index;not null;size:255"`
	MaxId    int64     `gorm:"not null;default:0"`
	Step     int64     `gorm:"not null;default:1024"`
	Desc     string    `gorm:"size:255"`
	UpdateAt time.Time `gorm:"autoUpdateTime"`
}
