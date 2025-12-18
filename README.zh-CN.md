# AI Code Reviewer (智能代码审查助手)

[![License](https://img.shields.io/badge/license-ISC-blue)](LICENSE)
[![TypeScript](https://img.shields.io/badge/language-TypeScript-blue)](https://www.typescriptlang.org/)
[![Docker](https://img.shields.io/badge/deployment-Docker-green)](Dockerfile)

**🇨🇳 中文** | [🇺🇸 English](./README.md)

一个基于 LLM (Large Language Model) 的自动化代码审查工具，支持多平台集成，能够像人类 Reviewer 一样提供精准的代码改进建议。

## ✨ 核心特性

- **多平台支持**: 完美适配 **GitHub**、**GitLab** 和 **Gitea**。
- **双运行模式**:
  - 🚀 **Webhook Server**: 部署为常驻服务，自动监听 PR/MR 事件。
  - 🛠 **CLI 工具**: 集成到 CI/CD 流水线 (GitHub Actions, GitLab CI) 中按需运行。
- **智能分析**:
  - 基于 OpenAI (GPT-4) 或兼容接口 (如 DeepSeek, Anthropic)。
  - **精准 Diff 解析**: 仅分析变更代码，自动映射行号，支持内联评论 (Inline Comments)。
  - **上下文感知**: 自动忽略锁文件、图片等非代码资源。
- **企业级特性**:
  - Docker 一键部署。
  - 完善的 Lint & Format 检查。
  - 可配置的 Prompt 和忽略规则。

## 🚀 快速开始

### 1. 前置要求

- Node.js 18+
- Docker (可选)
- OpenAI API Key (或兼容的 LLM Key)

### 2. 本地运行

```bash
# 安装依赖
npm install

# 配置环境变量
cp .env.example .env
# 编辑 .env 填入你的 Token

# 启动 Webhook 服务
npm run dev
```

### 3. Docker 部署

```bash
docker-compose up -d --build
```

服务将在 `http://localhost:4000` 启动。

## 📚 文档导航

- [📖 使用指南](docs/USAGE.zh-CN.md) - 详细配置与部署。
- [🏗 架构设计](docs/ARCHITECTURE.zh-CN.md) - 系统内部设计。
- [📝 需求规格](docs/REQUIREMENTS.zh-CN.md) - 功能需求列表。

## 🛠 开发与贡献

本项目包含完整的代码质量检查流程：

```bash
# 代码格式化
npm run format

# 静态检查
npm run lint
```

提交代码时会自动触发 Husky 钩子进行检查。

## 📄 License

ISC
