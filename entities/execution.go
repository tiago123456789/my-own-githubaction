package entities

import "gorm.io/gorm"

type Execution struct {
	gorm.Model
	ID        string `json:"id"`
	TriggerId uint   `json:"triggerId"`
	Status    string `json:"status"`
}
