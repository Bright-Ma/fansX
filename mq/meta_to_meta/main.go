package main

import (
	"bilibili/internal/model/mq"
	"encoding/json"
	"github.com/IBM/sarama"
	"log/slog"
	"time"
)

type Handler struct {
	producer *sarama.SyncProducer
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}
func (h *Handler) ConsumerClaim(session sarama.ConsumerGroupClaim, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		message := &mq.MetaToMetaMessage{}
		err := json.Unmarshal(msg.Value, message)
		if err != nil {
			slog.Error("unmarshal json:" + err.Error())
			continue
		}

	}
}
func main() {
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Retry.Backoff = time.Millisecond * 100
	config.Producer.RequiredAcks = sarama.WaitForAll

}
