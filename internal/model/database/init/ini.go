package main

import (
	"bilibili/internal/model/database"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	client, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	err = client.AutoMigrate(
		&database.Follower{},
		&database.Following{},
		&database.FollowerNums{},
		&database.FollowingNums{},
		&database.VisibleContentInfo{},
		&database.InvisibleContentInfo{},
	)
	if err != nil {
		panic(err.Error())
	}
}
