package segment

import (
	"gorm.io/gorm"
)

func (c *Creator) preApplication() {
	for {
		select {
		case <-c.ch:
			for {
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
	tx := c.db.Begin()
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
