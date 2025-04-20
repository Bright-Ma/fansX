package segment

import (
	"context"
	"gorm.io/gorm"
	"time"
)

func (c *Creator) GetId() (int64, bool) {
	c.mu.Lock()
	// 当前号段消耗殆尽并且新的号段还未申请到，防止过长时间的阻塞，故直接返回err
	if c.old.nextId == c.old.maxId+1 && c.new == nil {
		c.mu.Unlock()
		return 0, false
	}
	defer c.mu.Unlock()

	// 当前号段耗尽，更换为新号段
	if c.old.nextId == c.old.maxId+1 {
		c.old = c.new
		c.new = nil
	}
	//达到预申请阈值
	if c.old.nextId == c.old.preIndex {
		c.ch <- 1
	}
	res := c.old.nextId
	c.old.nextId++
	return res, true
}

func (c *Creator) GetIdWithContext(ctx context.Context) (int64, error) {
	var id int64
	var ok bool
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			id, ok = c.GetId()
			if !ok {
				time.Sleep(time.Millisecond * 50)
				continue
			}
		}
		break
	}
	return id, ctx.Err()
}

func (c *Creator) GetIdWithTimeout(timeout time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.GetIdWithContext(ctx)
}

// preApplication 向数据库预申请号段
func (c *Creator) preApplication() {
	for {
		select {
		case <-c.ch:
			for {
				// 无限尝试重试
				if err := c.tryApplication(); err != nil {
					continue
				} else {
					break
				}
			}
		}
	}
}

func (c *Creator) tryApplication() error {
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx := c.db.WithContext(timeout).Begin()
	err := tx.First(&IdTable{}, c.id).Update("MaxId", gorm.Expr("max_id + step")).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	record := IdTable{}
	err = tx.First(&record, c.id).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	c.mu.Lock()
	c.new = &buffer{
		nextId: record.MaxId - record.Step + 1,
		maxId:  record.MaxId,
	}
	c.new.preIndex = c.new.nextId + record.Step/10
	c.mu.Unlock()

	return nil
}
