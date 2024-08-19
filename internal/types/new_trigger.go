package types

type NewTrigger struct {
	WebhookUrl   string `json:"webhookUrl"`
	GithubSecret string `json:"secret"`
}
