package leaf_go

import (
	"errors"
	"leaf-go/segment"
)

type Config struct {
	Model         int
	SegmentConfig segment.Config
}

var (
	Segment   = 1
	Snowflake = 2
)

type Core interface {
	GetId() int64
}

func InitLeaf(c *Config) (Core, error) {
	if c.Model == Segment {
		return segment.NewCreator(&c.SegmentConfig)
	} else if c.Model == Snowflake {
		return nil, nil
	}

	return nil, errors.New("please select id model")
}
