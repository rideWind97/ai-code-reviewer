package config

import "os"

// Config 是整个服务启动时需要的配置集合。
//
// Go 项目里通常会把配置集中在一个结构体里，然后在 main 中传给各个模块。
// 这样业务代码不需要到处读取环境变量，测试时也更容易构造配置。
type Config struct {
	Addr        string       // HTTP 服务监听地址，例如 ":8080"。
	DatabaseURL string       // PostgreSQL 连接串；当前 MVP 先只读取，后续 store 会使用。
	GitHub      GitHubConfig // GitHub webhook 和 API 相关配置。
	Model       ModelConfig  // 模型网关相关配置。
}

// GitHubConfig 保存 GitHub provider 需要的配置。
type GitHubConfig struct {
	WebhookSecret string // GitHub webhook secret，用于校验请求签名。
	AppID         string // GitHub App ID；后续换 installation token 时使用。
	PrivateKey    string // GitHub App private key；后续换 installation token 时使用。
	Token         string // MVP 中直接调用 GitHub API 的 token。
}

// ModelConfig 保存 OpenAI-compatible 模型接口配置。
type ModelConfig struct {
	BaseURL string // 模型 API 地址，例如 "https://api.openai.com/v1"。
	Name    string // 模型名称，例如 "gpt-4.1" 或内部模型名。
	APIKey  string // 模型 API key。
}

// Load 从环境变量读取配置。
//
// 这里没有做强校验，是为了让测试和本地开发更轻量；真正部署时可以在 main
// 或启动脚本里要求某些关键环境变量必须存在。
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

// env 读取环境变量；如果变量为空，就返回 fallback。
func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
