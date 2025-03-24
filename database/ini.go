package main

import (
	"bilibili/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:@tcp(linux.1jian10.cn:4000)/relation?charset=utf8mb4&parseTime=True"
	client, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}
	err = client.AutoMigrate(&model.Follower{}, &model.Following{}, &model.FollowerNums{}, &model.FollowingNums{})
	if err != nil {
		panic(err.Error())
	}
}
