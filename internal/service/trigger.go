package service

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/tiago123456789/own-githubaction/internal/entities"
	"github.com/tiago123456789/own-githubaction/internal/repository"
	"github.com/tiago123456789/own-githubaction/internal/types"
	"github.com/tiago123456789/own-githubaction/pkg/file"
	"github.com/tiago123456789/own-githubaction/pkg/queue"
	secretmanager "github.com/tiago123456789/own-githubaction/pkg/secret_manager"
	"go.uber.org/zap"
)

type TriggerService struct {
	repository    repository.ITriggerRepository
	secretManager secretmanager.ISecretManager
	logger        *zap.Logger
	producer      queue.IProducer
	queueUtil     queue.IQueueUtil
	file          file.IFile
}

func NewTriggerService(
	secretManager secretmanager.ISecretManager,
	logger *zap.Logger, producer queue.IProducer,
	repository repository.ITriggerRepository,
	queueUtil queue.IQueueUtil,
	file file.IFile,

) *TriggerService {
	return &TriggerService{
		secretManager: secretManager,
		logger:        logger,
		producer:      producer,
		repository:    repository,
		queueUtil:     queueUtil,
		file:          file,
	}
}

func (t *TriggerService) GetTriggers() []entities.Trigger {
	return t.repository.FindAll()
}

func (t *TriggerService) GetExecutionsByTriggerId(triggerId string) []entities.Execution {
	return t.repository.FindExecutionsByTriggerId(triggerId)
}

func (t *TriggerService) GetExecutionLogsByTriggerIdAndExecutionId(
	triggerId string, exeuctionId string,
) []entities.ExecutionLog {
	return t.repository.GetExecutionLogsByTriggerIdAndExecutionId(
		triggerId, exeuctionId,
	)
}

func (t *TriggerService) Save(trigger types.Trigger) (string, error) {
	hasEnvs := len(trigger.Envs) > 0
	triggerToSave := &entities.Trigger{
		Hash:            trigger.Hash,
		ActionToRun:     trigger.ActionToRun,
		LinkRepository:  trigger.LinkRepository,
		RepositoryToken: trigger.RepositoryToken,
		IsPrivate:       trigger.IsPrivate,
		HasEnvs:         hasEnvs,
	}

	t.repository.Save(triggerToSave)

	if hasEnvs {
		envsJSON, _ := json.Marshal(trigger.Envs)
		err := t.secretManager.Add(trigger.Hash, string(envsJSON))
		if err != nil {
			t.logger.Error(
				fmt.Sprintf("Failed to create secret: %v", err),
			)

			return "", errors.New("Internal server error")
		}
	}

	apiBaseUrl := os.Getenv("API_BASE_URL")
	return fmt.Sprintf("%s/triggers-execute/%s", apiBaseUrl, trigger.Hash), nil
}

func (t *TriggerService) Execute(hash string) (entities.Execution, error) {
	trigger := t.repository.FindByHash(hash)

	if trigger.ID == 0 {
		return entities.Execution{}, errors.New("Not found register")
	}

	execution := entities.Execution{}
	execution.Status = "Queued"
	execution.ID = uuid.NewString()
	execution.TriggerId = trigger.ID

	t.repository.SaveExecution(&execution)

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

	t.producer.Publish(executionMessage)
	return execution, nil
}

func (t *TriggerService) getEnvsDotenvFileFormat(hash string) (string, error) {
	secret, err := t.secretManager.Get(hash)
	if err != nil {
		t.logger.Error(
			fmt.Sprintf("Failed to get secret: %v", err),
		)

		return "", err
	}

	envs := ""
	var secretsFromSecretManager map[string]string
	json.Unmarshal([]byte(secret), &secretsFromSecretManager)
	for key, value := range secretsFromSecretManager {
		envs += fmt.Sprintf("%s='%s'\n", key, value)
	}

	return envs, nil
}

func (t *TriggerService) ProcessPipeline(payload []byte) error {
	p := types.Execution{}
	err := t.queueUtil.ParseMessage(payload, &p)
	if err != nil {
		t.logger.Error(
			fmt.Sprintf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry),
		)
		return err
	}

	fileName := fmt.Sprintf("pipelines/.env.%s", p.ID)

	if p.Trigger.HasEnvs {
		envs, err := t.getEnvsDotenvFileFormat(p.Trigger.Hash)
		if err != nil {
			t.logger.Error(
				fmt.Sprintf("Failed to get secret: %v", err),
			)

			return err
		}

		err = t.file.WriteFile(fileName, envs)
		if err != nil {
			t.logger.Error(
				fmt.Sprintf("Error writing to file: %v", err),
			)
			return err
		}

	}

	t.logger.Info(
		fmt.Sprintf(
			"Start to process exection with id %s the project %s pipeline %s",
			p.ID,
			p.Trigger.LinkRepository,
			p.Trigger.ActionToRun,
		),
	)

	execution := t.repository.FindExecutionById(p.ID)
	t.repository.UpdateExecutionData(&execution, entities.Execution{Status: "In Progress"})

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
		paramsCommand = append(paramsCommand, fmt.Sprintf(".env.%s", p.ID))
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
		t.repository.SaveExecutionLog(&executionLog)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading output:", err)
	}

	if err := cmd.Wait(); err != nil {
		t.logger.Error(
			fmt.Sprintf(
				"The process exection with id %s the project %s pipeline %s was failed. Caused by: %s",
				p.ID,
				p.Trigger.LinkRepository,
				p.Trigger.ActionToRun,
				err.Error(),
			),
		)
		t.repository.UpdateExecutionData(&execution, entities.Execution{Status: "Failed"})
	} else {
		t.logger.Info(
			fmt.Sprintf(
				"The process exection with id %s the project %s pipeline %s is Done",
				p.ID,
				p.Trigger.LinkRepository,
				p.Trigger.ActionToRun,
			),
		)
		t.repository.UpdateExecutionData(&execution, entities.Execution{Status: "Done"})
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
		t.logger.Error(
			"Error:",
			zap.Error(fmt.Errorf("%w", err)),
		)
	}

	t.logger.Info(
		fmt.Sprintf(
			"Finished to process exection with id %s the project %s pipeline %s",
			p.ID,
			p.Trigger.LinkRepository,
			p.Trigger.ActionToRun,
		),
	)
	return nil
}
