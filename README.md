# AI Code Reviewer

[![License](https://img.shields.io/badge/license-ISC-blue)](LICENSE)
[![TypeScript](https://img.shields.io/badge/language-TypeScript-blue)](https://www.typescriptlang.org/)
[![Docker](https://img.shields.io/badge/deployment-Docker-green)](Dockerfile)

[🇨🇳 中文](./README.zh-CN.md) | **🇺🇸 English**

An automated code review tool powered by LLMs (Large Language Models), supporting multiple platforms. It provides precise, human-like code improvement suggestions.

## ✨ Core Features

- **Multi-Platform Support**: Seamless integration with **GitHub**, **GitLab**, and **Gitea**.
- **Dual Operation Modes**:
  - 🚀 **Webhook Server**: Deployed as a persistent service to automatically listen for PR/MR events.
  - 🛠 **CLI Tool**: Integrated into CI/CD pipelines (GitHub Actions, GitLab CI) for on-demand execution.
- **Intelligent Analysis**:
  - Powered by OpenAI (GPT-4) or compatible APIs (DeepSeek, Anthropic).
  - **Precise Diff Parsing**: Analyzes only changed code, maps line numbers automatically, supports Inline Comments.
  - **Context Awareness**: Automatically ignores lock files, images, and non-code assets.
- **Enterprise Ready**:
  - Docker deployment ready.
  - Comprehensive Lint & Format checks.
  - Configurable Prompts and ignore rules.

## 🚀 Quick Start

### 1. Prerequisites

- Node.js 18+
- Docker (Optional)
- OpenAI API Key (or compatible LLM Key)

### 2. Local Execution

```bash
# Install dependencies
npm install

# Configure Environment Variables
cp .env.example .env
# Edit .env and fill in your Tokens

# Start Webhook Server
npm run dev
```

### 3. Docker Deployment

```bash
docker-compose up -d --build
```

The service will start at `http://localhost:3000`.

## 📚 Documentation

- [📖 Usage Guide](docs/USAGE.md) - Configuration & Deployment.
- [🏗 Architecture](docs/ARCHITECTURE.md) - System Design.
- [📝 Requirements](docs/REQUIREMENTS.md) - Feature List.

## 🛠 Development

This project includes a complete code quality check process:

```bash
# Format Code
npm run format

# Static Analysis
npm run lint
```

Husky hooks will be triggered automatically upon commit.

## 📄 License

ISC
