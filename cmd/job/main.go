package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/tiago123456789/own-githubaction/internal/config"
	"github.com/tiago123456789/own-githubaction/internal/repository"
	"github.com/tiago123456789/own-githubaction/internal/service"
	"github.com/tiago123456789/own-githubaction/pkg/file"
	"github.com/tiago123456789/own-githubaction/pkg/logger"
	"github.com/tiago123456789/own-githubaction/pkg/queue"
	secretmanager "github.com/tiago123456789/own-githubaction/pkg/secret_manager"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := config.GetDB()
	logger := logger.Get()
	secretManager := secretmanager.New(true)

	producerQueue := queue.NewProducer("pipeline_executions")
	defer producerQueue.Close()

	triggerRepository := repository.NewTriggerRepository(db)
	triggerService := service.NewTriggerService(
		secretManager,
		logger, producerQueue,
		triggerRepository,
		queue.NewQueueUtil(),
		file.New(logger),
	)

	consumerQueue := queue.NewConsumer(
		"pipeline_executions",
		triggerService.ProcessPipeline,
	)

	consumerQueue.Listen()
}
