package entities

import "gorm.io/gorm"

type Trigger struct {
	gorm.Model
	Hash            string `json:"hash"`
	ActionToRun     string `json:"actionToRun"`
	LinkRepository  string `json:"linkRepository"`
	IsPrivate       bool   `json:"isPrivate"`
	RepositoryToken string `json:"repositoryToken"`
	HasEnvs         bool   `json:"hasEnvs"`
}
