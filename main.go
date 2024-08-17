package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/tiago123456789/own-githubaction/entities"
	"github.com/tiago123456789/own-githubaction/logger"
	"github.com/tiago123456789/own-githubaction/queue"
	secretmanager "github.com/tiago123456789/own-githubaction/secret_manager"
	"github.com/tiago123456789/own-githubaction/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := gorm.Open(sqlite.Open("db.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(
		&entities.Trigger{}, &entities.Execution{},
		&entities.ExecutionLog{},
	)

	producerQueue := queue.NewProducer("pipeline_executions")
	defer producerQueue.Close()

	secretManager := secretmanager.New(true)
	logger := logger.Get()

	app := fiber.New()

	hasAuthorization := func(c *fiber.Ctx) error {
		apiKey := os.Getenv("API_KEY")
		if apiKey != c.Get("x-api-key") {
			return c.Status(403).JSON(fiber.Map{
				"message": "You don't have permission to do that action",
			})
		}

		return c.Next()
	}

	app.Post("/triggers-execute/:hash", func(c *fiber.Ctx) error {
		var trigger entities.Trigger

		db.First(&trigger, "hash = ?", c.Params("hash"))

		if trigger.ID == 0 {
			return c.Status(404).JSON(fiber.Map{
				"message": "Not found register",
			})
		}

		execution := entities.Execution{}

		execution.Status = "Queued"
		execution.ID = uuid.NewString()
		execution.TriggerId = trigger.ID

		db.Create(&execution)

		executionMessage := types.Execution{
			ID:        execution.ID,
			TriggerId: int(trigger.ID),
			Status:    execution.Status,
			Trigger: types.Trigger{
				ID:              int(trigger.ID),
				Hash:            trigger.Hash,
				ActionToRun:     trigger.ActionToRun,
				LinkRepository:  fmt.Sprint(trigger.LinkRepository, ".git"),
				IsPrivate:       trigger.IsPrivate,
				RepositoryToken: trigger.RepositoryToken,
				HasEnvs:         trigger.HasEnvs,
			},
		}

		producerQueue.Publish(executionMessage)
		return c.JSON(execution)
	})

	app.Get("/triggers/:id/executions", hasAuthorization, func(c *fiber.Ctx) error {
		var executions []entities.Execution
		db.Order("created_at desc").Find(&executions, "trigger_id = ?", c.Params("id"))
		return c.JSON(executions)
	})

	app.Get("/triggers/:id/executions/:executionId/logs", hasAuthorization, func(c *fiber.Ctx) error {
		var executionsLogs []entities.ExecutionLog
		db.Order("created_at asc").Find(&executionsLogs, "execution_id = ?", c.Params("executionId"))
		return c.JSON(executionsLogs)
	})

	app.Get("/triggers", hasAuthorization, func(c *fiber.Ctx) error {
		var registers []entities.Trigger
		db.Find(&registers)
		return c.JSON(registers)
	})

	app.Post("/triggers", hasAuthorization, func(c *fiber.Ctx) error {
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

		hasEnvs := len(trigger.Envs) > 0
		triggerToSave := &entities.Trigger{
			Hash:            trigger.Hash,
			ActionToRun:     trigger.ActionToRun,
			LinkRepository:  trigger.LinkRepository,
			RepositoryToken: trigger.RepositoryToken,
			IsPrivate:       trigger.IsPrivate,
			HasEnvs:         hasEnvs,
		}

		db.Create(triggerToSave)

		if hasEnvs {
			envsJSON, _ := json.Marshal(trigger.Envs)
			err := secretManager.Add(trigger.Hash, string(envsJSON))
			if err != nil {
				logger.Error(
					fmt.Sprintf("Failed to create secret: %v", err),
				)

				return c.Status(500).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}

		apiBaseUrl := os.Getenv("API_BASE_URL")
		return c.JSON(types.NewTrigger{
			Url: fmt.Sprintf("%s/triggers-execute/%s", apiBaseUrl, trigger.Hash),
		})
	})

	app.Listen(":3000")
}
