package main

import (
	"fansX/internal/model/database"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	err = db.Take(&database.VisibleContentInfo{}, 1924756313644990464).Updates(map[string]interface{}{
		"status":  database.ContentStatusDelete,
		"version": 4,
	}).Error
	if err != nil {
		panic(err.Error())
	}
	return

}
