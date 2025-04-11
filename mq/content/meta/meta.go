package main

import (
	"bilibili/common/util"
	"context"
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

	logger, err := util.InitLog("mq:MetaContent", slog.LevelDebug)
	if err != nil {
		panic(err.Error())
	}
	slog.SetDefault(logger)

	consumer, _ := sarama.NewConsumerGroup([]string{"1jian10.cn:9094"}, "test_meta_content_group", config)
	handler := Handler{
		db: db,
	}

	err = consumer.Consume(context.Background(), []string{"test_meta_content"}, &handler)
	if err != nil {
		panic(err.Error())
	}

}
