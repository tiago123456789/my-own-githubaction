package queue

import (
	"context"
	"log"
	"os"

	"github.com/hibiken/asynq"
)

type IConsumer interface {
	Listen()
}

type Handler func([]byte) error

type Consumer struct {
	client    *asynq.Server
	mux       *asynq.ServeMux
	queueName string
}

func NewConsumer(queueName string, handler Handler) *Consumer {
	redisAddr := os.Getenv("REDIS_URL")

	client := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 1,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queueName, func(ctx context.Context, t *asynq.Task) error {
		error := handler(t.Payload())
		return error
	})

	return &Consumer{
		queueName: queueName,
		client:    client,
		mux:       mux,
	}
}

func (c *Consumer) Listen() {
	if err := c.client.Run(c.mux); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
