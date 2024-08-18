package repository

import (
	"github.com/tiago123456789/own-githubaction/internal/entities"
	"gorm.io/gorm"
)

type ITriggerRepository interface {
	FindAll() []entities.Trigger
	FindExecutionsByTriggerId(triggerId string) []entities.Execution
	GetExecutionLogsByTriggerIdAndExecutionId(
		triggerId string, exeuctionId string,
	) []entities.ExecutionLog
	Save(data *entities.Trigger)
	FindByHash(hash string) entities.Trigger
	SaveExecution(data *entities.Execution)
	FindExecutionById(id string) entities.Execution
	UpdateExecutionData(
		execution *entities.Execution, dataModified entities.Execution,
	)
	SaveExecutionLog(executionLog *entities.ExecutionLog)
}

type TriggerRepository struct {
	db *gorm.DB
}

func NewTriggerRepository(
	db *gorm.DB,
) *TriggerRepository {
	return &TriggerRepository{
		db: db,
	}
}

func (t *TriggerRepository) FindAll() []entities.Trigger {
	var registers []entities.Trigger
	t.db.Find(&registers)

	return registers
}

func (t *TriggerRepository) FindExecutionsByTriggerId(triggerId string) []entities.Execution {
	var executions []entities.Execution
	t.db.Order("created_at desc").Find(&executions, "trigger_id = ?", triggerId)
	return executions
}

func (t *TriggerRepository) GetExecutionLogsByTriggerIdAndExecutionId(
	triggerId string, exeuctionId string,
) []entities.ExecutionLog {
	var executionsLogs []entities.ExecutionLog
	t.db.Order("created_at asc").Find(&executionsLogs, "execution_id = ?", exeuctionId)
	return executionsLogs
}

func (t *TriggerRepository) Save(data *entities.Trigger) {
	t.db.Create(data)
}

func (t *TriggerRepository) FindByHash(hash string) entities.Trigger {
	var trigger entities.Trigger

	t.db.First(&trigger, "hash = ?", hash)

	return trigger
}

func (t *TriggerRepository) SaveExecution(data *entities.Execution) {
	t.db.Create(data)
}

func (t *TriggerRepository) FindExecutionById(id string) entities.Execution {
	var execution entities.Execution
	t.db.Find(&execution, "id = ?", id)
	return execution
}

func (t *TriggerRepository) UpdateExecutionData(
	execution *entities.Execution, dataModified entities.Execution,
) {
	t.db.Model(execution).Updates(dataModified)
}

func (t *TriggerRepository) SaveExecutionLog(executionLog *entities.ExecutionLog) {
	t.db.Save(&executionLog)
}
