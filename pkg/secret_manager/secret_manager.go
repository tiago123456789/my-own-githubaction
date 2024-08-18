package secretmanager

import (
	"log"
	"os"

	"github.com/phasehq/golang-sdk/phase"
)

type ISecretManager interface {
	Add(key string, value string) error
	Get(key string) (string, error)
}

type SecretManager struct {
	client *phase.Phase
}

func New(enableDebug bool) *SecretManager {
	phaseTokenService := os.Getenv("PHASE_TOKEN_SERVICE")
	host := os.Getenv("PHASE_HOST")
	phaseClient := phase.Init(phaseTokenService, host, enableDebug)

	if phaseClient == nil {
		log.Fatal("Failed to initialize Phase client")
	}

	return &SecretManager{
		client: phaseClient,
	}
}

func (s *SecretManager) Add(key string, value string) error {
	appName := os.Getenv("PHASE_PROJECT")
	envName := os.Getenv("PHASE_ENV")

	opts := phase.CreateSecretsOptions{
		KeyValuePairs: []map[string]string{
			{(key): value},
		},
		EnvName: envName,
		AppName: appName,
	}

	err := s.client.Create(opts)

	if err != nil {
		return err
	}

	return nil
}

func (s *SecretManager) Get(key string) (string, error) {
	appName := os.Getenv("PHASE_PROJECT")
	envName := os.Getenv("PHASE_ENV")

	opts := phase.GetSecretOptions{
		KeyToFind: key,
		EnvName:   envName,
		AppName:   appName,
	}

	secret, err := s.client.Get(opts)
	if err != nil {
		return "", err
	}

	value := (*secret)["value"].(string)
	return value, nil
}
