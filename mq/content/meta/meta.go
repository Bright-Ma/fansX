package main

import (
	"context"
	"fansX/internal/model/database"
	"fansX/internal/util"
	"fmt"
	"github.com/IBM/sarama"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
)

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.AutoCommit.Enable = false

	dsn := "root:@tcp(linux.1jian10.cn:4000)/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err.Error())
	}

	logger, err := util.InitLog("metacontent.kafka", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}
	slog.SetDefault(logger)

	factory := NewFactory()
	factory.RegisterStrategy(database.ContentStatusPass, &PassStrategy{})
	factory.RegisterStrategy(database.ContentStatusDelete, &DeleteStrategy{})
	factory.RegisterStrategy(database.ContentStatusCheck, &CheckStrategy{})
	factory.RegisterStrategy(database.ContentStatusNotPass, &NotPassStrategy{})

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_content_meta_group", config)
	handler := Handler{
		db:      db,
		factory: factory,
	}
	fmt.Println("in")
	err = consumer.Consume(context.Background(), []string{"test_content_meta"}, &handler)
	if err != nil {
		panic(err.Error())
	}
}
