package queue

import (
	"encoding/json"
	"os"

	"github.com/hibiken/asynq"
)

type Producer struct {
	client    *asynq.Client
	queueName string
}

func NewProducer(queueName string) *Producer {

	redisAddr := os.Getenv("REDIS_URL")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	return &Producer{
		client:    client,
		queueName: queueName,
	}
}

func (p *Producer) Publish(payload interface{}) {
	payloadSendQeueue, _ := json.Marshal(payload)

	p.client.Enqueue(
		asynq.NewTask(p.queueName, payloadSendQeueue),
	)
}

func (p *Producer) Close() {
	p.client.Close()
}
