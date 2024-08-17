package types

type Trigger struct {
	ID              int               `json: "id"`
	Hash            string            `json:"hash"`
	ActionToRun     string            `json:"actionToRun"`
	LinkRepository  string            `json:"linkRepository"`
	IsPrivate       bool              `json:"isPrivate"`
	RepositoryToken string            `json:"repositoryToken"`
	Envs            map[string]string `json:"envs"`
	HasEnvs         bool              `json:"hasEnvs"`
}
