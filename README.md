# AI Code Reviewer / 智能代码审查助手

![License](https://img.shields.io/badge/license-ISC-blue)
![TypeScript](https://img.shields.io/badge/language-TypeScript-blue)
![Docker](https://img.shields.io/badge/deployment-Docker-green)

**English**: An automated code review tool powered by LLMs (Large Language Models), supporting multiple platforms. It provides precise, human-like code improvement suggestions.

**中文**: 一个基于 LLM (Large Language Model) 的自动化代码审查工具，支持多平台集成，能够像人类 Reviewer 一样提供精准的代码改进建议。

## ✨ Core Features / 核心特性

- **Multi-Platform Support / 多平台支持**:
  - Seamless integration with **GitHub**, **GitLab**, and **Gitea**.
  - 完美适配 **GitHub**、**GitLab** 和 **Gitea**。

- **Dual Operation Modes / 双运行模式**:
  - 🚀 **Webhook Server**: Deployed as a persistent service to automatically listen for PR/MR events.
  - 🚀 **Webhook Server**: 部署为常驻服务，自动监听 PR/MR 事件。
  - 🛠 **CLI Tool**: Integrated into CI/CD pipelines (GitHub Actions, GitLab CI) for on-demand execution.
  - 🛠 **CLI 工具**: 集成到 CI/CD 流水线 (GitHub Actions, GitLab CI) 中按需运行。

- **Intelligent Analysis / 智能分析**:
  - Powered by OpenAI (GPT-4) or compatible APIs (DeepSeek, Anthropic).
  - 基于 OpenAI (GPT-4) 或兼容接口 (如 DeepSeek, Anthropic)。
  - **Precise Diff Parsing**: Analyzes only changed code, maps line numbers automatically, supports Inline Comments.
  - **精准 Diff 解析**: 仅分析变更代码，自动映射行号，支持内联评论 (Inline Comments)。
  - **Context Awareness**: Automatically ignores lock files, images, and non-code assets.
  - **上下文感知**: 自动忽略锁文件、图片等非代码资源。

- **Enterprise Ready / 企业级特性**:
  - Docker deployment ready. / Docker 一键部署。
  - Comprehensive Lint & Format checks. / 完善的 Lint & Format 检查。
  - Configurable Prompts and ignore rules. / 可配置的 Prompt 和忽略规则。

## 🚀 Quick Start / 快速开始

### 1. Prerequisites / 前置要求

- Node.js 18+
- Docker (Optional / 可选)
- OpenAI API Key (or compatible LLM Key / 或兼容的 LLM Key)

### 2. Local Execution / 本地运行

```bash
# Install dependencies / 安装依赖
npm install

# Configure Environment Variables / 配置环境变量
cp .env.example .env
# Edit .env and fill in your Tokens / 编辑 .env 填入你的 Token

# Start Webhook Server / 启动 Webhook 服务
npm run dev
```

### 3. Docker Deployment / Docker 部署

```bash
docker-compose up -d --build
```

The service will start at `http://localhost:3000`.
服务将在 `http://localhost:3000` 启动。

## 📚 Documentation / 文档导航

- [📖 Usage Guide / 使用指南](docs/USAGE.md) - Configuration & Deployment / 详细配置与部署。
- [🏗 Architecture / 架构设计](docs/ARCHITECTURE.md) - System Design / 系统内部设计。
- [📝 Requirements / 需求规格](docs/REQUIREMENTS.md) - Feature List / 功能需求列表。

## 🛠 Development / 开发与贡献

This project includes a complete code quality check process:
本项目包含完整的代码质量检查流程：

```bash
# Format Code / 代码格式化
npm run format

# Static Analysis / 静态检查
npm run lint
```

Husky hooks will be triggered automatically upon commit.
提交代码时会自动触发 Husky 钩子进行检查。

## 📄 License

ISC
