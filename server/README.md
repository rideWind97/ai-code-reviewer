# AI Code Reviewer Server

## 配置

本项目支持多种 AI 提供商，包括 OpenAI 兼容接口（如 DeepSeek, Qwen）和 Google Gemini。

### 环境变量设置

在 `.env` 文件中配置以下变量：

```bash
# AI Provider (openai, deepseek, gemini)
AI_PROVIDER=gemini

# API Key
AI_API_KEY=your_api_key_here

# Model Name
# For Gemini: gemini-1.5-pro, gemini-1.5-flash, gemini-pro
# For OpenAI/DeepSeek: deepseek-chat, qwen-plus, gpt-4
AI_MODEL=gemini-1.5-pro

# API URL (Optional, for OpenAI compatible providers only)
# AI_API_URL=https://api.deepseek.com/v1
```

### Google Gemini 配置

1. 获取 Google AI Studio API Key: [Google AI Studio](https://aistudio.google.com/)
2. 设置 `AI_PROVIDER=gemini`
3. 设置 `AI_API_KEY` 为你的 Gemini API Key
4. 设置 `AI_MODEL` (默认为 `gemini-1.5-pro`)

### DeepSeek / OpenAI 配置

1. 设置 `AI_PROVIDER=deepseek` 或 `openai`
2. 设置 `AI_API_KEY`
3. 设置 `AI_API_URL` (例如 DeepSeek: `https://api.deepseek.com`)
4. 设置 `AI_MODEL` (例如 `deepseek-chat`)

