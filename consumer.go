package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/tiago123456789/own-githubaction/entities"
	"github.com/tiago123456789/own-githubaction/logger"
	"github.com/tiago123456789/own-githubaction/queue"
	secretmanager "github.com/tiago123456789/own-githubaction/secret_manager"
	"github.com/tiago123456789/own-githubaction/types"
	"go.uber.org/zap"
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

	logger := logger.Get()
	secretManager := secretmanager.New(true)

	consumerQueue := queue.NewConsumer(
		"pipeline_executions",
		func(payload []byte) error {
			var p types.Execution
			if err := json.Unmarshal(payload, &p); err != nil {
				logger.Error(
					fmt.Sprintf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry),
				)
				return err
			}

			fileName := fmt.Sprintf("pipelines/.env.%s", p.ID)

			if p.Trigger.HasEnvs {
				secret, err := secretManager.Get(p.Trigger.Hash)
				if err != nil {
					logger.Error(
						fmt.Sprintf("Failed to get secret: %v", err),
					)

					return err
				}

				envs := ""
				var secretsFromSecretManager map[string]string
				json.Unmarshal([]byte(secret), &secretsFromSecretManager)
				for key, value := range secretsFromSecretManager {
					envs += fmt.Sprintf("%s='%s'\n", key, value)
				}

				file, err := os.Create(fileName)
				defer file.Close()

				if err != nil {
					logger.Error(
						fmt.Sprintf("Error creating file: %v", err),
					)
					return err
				}

				_, err = file.WriteString(envs)
				if err != nil {
					logger.Error(
						fmt.Sprintf("Error writing to file: %v", err),
					)
					return err
				}

			}

			logger.Info(
				fmt.Sprintf(
					"Start to process exection with id %s the project %s pipeline %s",
					p.ID,
					p.Trigger.LinkRepository,
					p.Trigger.ActionToRun,
				),
			)
			var execution entities.Execution
			db.Find(&execution, "id = ?", p.ID)
			db.Model(&execution).Updates(entities.Execution{Status: "In Progress"})

			if p.Trigger.IsPrivate {
				repositoryLinkSplited := strings.Split(p.Trigger.LinkRepository, "/")
				githubUser := repositoryLinkSplited[3]
				repository := repositoryLinkSplited[len(repositoryLinkSplited)-1]
				p.Trigger.LinkRepository = fmt.Sprintf("%s//%s:%s@github.com/%s/%s",
					repositoryLinkSplited[0], githubUser, p.Trigger.RepositoryToken, githubUser, repository,
				)
			}

			commandsToExecute := "cd pipelines && git clone %s %s && cd %s && act -W '.github/workflows/%s'"
			paramsCommand := []interface{}{
				p.Trigger.LinkRepository,
				p.ID,
				p.ID,
				p.Trigger.ActionToRun,
			}

			if p.Trigger.HasEnvs {
				commandsToExecute += " --secret-file ../%s"
				paramsCommand = append(paramsCommand, fileName)
			}

			cmd := exec.Command(
				"bash", "-c",
				fmt.Sprintf(
					commandsToExecute,
					paramsCommand...,
				),
			)

			stdoutPipe, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Println("Error:", err)
				return nil
			}

			if err := cmd.Start(); err != nil {
				fmt.Println("Error:", err)
				return nil
			}

			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Text()
				executionLog := entities.ExecutionLog{}
				executionLog.ExecutionId = p.ID
				executionLog.Log = line
				executionLog.ID = uuid.NewString()
				db.Save(&executionLog)
			}

			if err := scanner.Err(); err != nil {
				fmt.Println("Error reading output:", err)
			}

			if err := cmd.Wait(); err != nil {
				logger.Error(
					fmt.Sprintf(
						"The process exection with id %s the project %s pipeline %s was failed. Caused by: %s",
						p.ID,
						p.Trigger.LinkRepository,
						p.Trigger.ActionToRun,
						err.Error(),
					),
				)
				db.Model(&execution).Updates(entities.Execution{Status: "Failed"})
			} else {
				logger.Info(
					fmt.Sprintf(
						"The process exection with id %s the project %s pipeline %s is Done",
						p.ID,
						p.Trigger.LinkRepository,
						p.Trigger.ActionToRun,
					),
				)
				db.Model(&execution).Updates(entities.Execution{Status: "Done"})
			}

			finalCommandsToExecute := "cd pipelines && rm -r %s"
			paramsFinalCommand := []interface{}{
				p.ID,
			}

			if p.Trigger.HasEnvs {
				finalCommandsToExecute += " && cd .. && rm -f ./%s"
				paramsFinalCommand = append(paramsFinalCommand, fileName)
			}

			cmd = exec.Command(
				"bash", "-c",
				fmt.Sprintf(
					finalCommandsToExecute,
					paramsFinalCommand...,
				),
			)

			_, err = cmd.CombinedOutput()
			if err != nil {
				logger.Error(
					"Error:",
					zap.Error(fmt.Errorf("%w", err)),
				)
			}

			logger.Info(
				fmt.Sprintf(
					"Finished to process exection with id %s the project %s pipeline %s",
					p.ID,
					p.Trigger.LinkRepository,
					p.Trigger.ActionToRun,
				),
			)
			return nil
		},
	)

	consumerQueue.Listen()
}
