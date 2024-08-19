package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/tiago123456789/own-githubaction/internal/config"
	"github.com/tiago123456789/own-githubaction/internal/entities"
	"github.com/tiago123456789/own-githubaction/internal/middleware"
	"github.com/tiago123456789/own-githubaction/internal/repository"
	"github.com/tiago123456789/own-githubaction/internal/service"
	"github.com/tiago123456789/own-githubaction/internal/types"
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
	db.AutoMigrate(
		&entities.Trigger{}, &entities.Execution{},
		&entities.ExecutionLog{},
	)

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

	app := fiber.New()

	app.Post("/triggers-execute/:hash", middleware.HasValidSecret, func(c *fiber.Ctx) error {
		execution, err := triggerService.Execute(c.Params("hash"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"message": "Not found register",
			})
		}

		return c.JSON(execution)
	})

	app.Get("/triggers/:id/executions", middleware.HasAuthorization, func(c *fiber.Ctx) error {
		return c.JSON(triggerService.GetExecutionsByTriggerId(c.Params("id")))
	})

	app.Get("/triggers/:id/executions/:executionId/logs", middleware.HasAuthorization, func(c *fiber.Ctx) error {
		return c.JSON(triggerService.GetExecutionLogsByTriggerIdAndExecutionId(
			c.Params("id"),
			c.Params("executionId"),
		))
	})

	app.Get("/triggers", middleware.HasAuthorization, func(c *fiber.Ctx) error {
		return c.JSON(triggerService.GetTriggers())
	})

	app.Post("/triggers", middleware.HasAuthorization, func(c *fiber.Ctx) error {
		trigger := &types.Trigger{}
		if err := c.BodyParser(trigger); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if len(trigger.ActionToRun) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"message": "The field actionToRun is required",
			})
		}

		if len(trigger.LinkRepository) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"message": "The field linkRepository is required",
			})
		}

		if trigger.IsPrivate == true && len(trigger.RepositoryToken) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"message": "When repository is private the field repositoryToken is required",
			})
		}

		trigger.Hash = uuid.NewString()

		webhookUrl, err := triggerService.Save(*trigger)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		return c.JSON(types.NewTrigger{
			WebhookUrl:   webhookUrl,
			GithubSecret: trigger.Hash,
		})
	})

	app.Listen(":3000")
}
