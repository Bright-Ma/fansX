package leaf_go

import (
	"bilibili/common/middleware/leaf-go/segment"
	"bilibili/common/middleware/leaf-go/snowflake"
	"context"
	"errors"
	"time"
)

type Config struct {
	Model           int
	SegmentConfig   *segment.Config
	SnowflakeConfig *snowflake.Config
}

var (
	Segment   = 1
	Snowflake = 2
)

type Core interface {
	GetId() int64
}

func Init(c *Config) (Core, error) {
	if c.Model == Segment {
		return segment.NewCreator(c.SegmentConfig)
	} else if c.Model == Snowflake {
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		return snowflake.NewCreator(timeout, c.SnowflakeConfig)
	}

	return nil, errors.New("please select id model")
}
