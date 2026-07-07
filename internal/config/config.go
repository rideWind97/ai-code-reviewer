package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
	GitHub      GitHubConfig
	Model       ModelConfig
}

type GitHubConfig struct {
	WebhookSecret string
	AppID         string
	PrivateKey    string
	Token         string
}

type ModelConfig struct {
	BaseURL string
	Name    string
	APIKey  string
}

func Load() (Config, error) {
	return Config{
		Addr:        env("AICR_ADDR", ":8080"),
		DatabaseURL: os.Getenv("AICR_DATABASE_URL"),
		GitHub: GitHubConfig{
			WebhookSecret: os.Getenv("AICR_GITHUB_WEBHOOK_SECRET"),
			AppID:         os.Getenv("AICR_GITHUB_APP_ID"),
			PrivateKey:    os.Getenv("AICR_GITHUB_PRIVATE_KEY"),
			Token:         os.Getenv("AICR_GITHUB_TOKEN"),
		},
		Model: ModelConfig{
			BaseURL: os.Getenv("AICR_MODEL_BASE_URL"),
			Name:    os.Getenv("AICR_MODEL_NAME"),
			APIKey:  os.Getenv("AICR_MODEL_API_KEY"),
		},
	}, nil
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
