package entities

import "gorm.io/gorm"

type ExecutionLog struct {
	gorm.Model
	ID          string `json:"id"`
	ExecutionId string `json:"executionId"`
	Log         string `json:"log"`
}
