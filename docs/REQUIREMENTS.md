# Product Requirements Document (PRD) / 产品需求文档

## 1. Project Overview / 项目概览

**English**: A Node.js-based tool that utilizes Large Language Models (LLM) to automatically review code changes (Pull Requests/Merge Requests) across multiple Git hosting platforms. It aims to catch potential bugs, style issues, and security vulnerabilities before human review.

**中文**: 一个基于 Node.js 的工具，利用大语言模型 (LLM) 自动审查多个 Git 托管平台上的代码变更 (PR/MR)。旨在人工审查之前捕捉潜在的 Bug、样式问题和安全漏洞。

## 2. Key Features / 核心特性

### 2.1 Multi-Platform Support / 多平台支持

The system must support the following platforms via a unified abstraction layer:
系统必须通过统一的抽象层支持以下平台：

- **GitHub** (Cloud & Enterprise)
- **GitLab** (Cloud & Self-hosted)
- **Gitea** (Self-hosted)

### 2.2 Integration Modes / 集成模式

- **Webhook Server Mode**: Listens for `pull_request` or `merge_request` events and triggers reviews automatically.
- **Webhook Server 模式**: 监听 `pull_request` 或 `merge_request` 事件并自动触发审查。
- **CI/CD CLI Mode**: Can be run as a step in a pipeline (e.g., GitHub Actions, GitLab CI) to review the current commit/PR.
- **CI/CD CLI 模式**: 可作为流水线步骤运行，审查当前提交/PR。

### 2.3 Code Analysis Capabilities / 代码分析能力

- **Diff Analysis**: Fetches the diff of the PR, parsing added/modified files.
- **Diff 分析**: 获取 PR 的差异，解析新增/修改的文件。
- **Context Awareness**: Understands file types and provides language-specific advice.
- **上下文感知**: 理解文件类型并提供特定语言的建议。
- **Filtering**: Ignores generated files (e.g., `package-lock.json`, `dist/`), binary files, or specific patterns defined in configuration.
- **过滤**: 忽略生成文件、二进制文件或配置中定义的特定模式。

### 2.4 Feedback Mechanism / 反馈机制

- **Inline Comments**: Specific comments on the exact line where the issue is detected.
- **内联评论**: 在检测到问题的具体行上发表评论。
- **General Summary**: A top-level comment summarizing the overall quality and security score of the PR.
- **总体总结**: 对 PR 的整体质量和安全性评分进行顶层评论总结。

## 3. Configuration & Customization / 配置与自定义

Users should be able to configure via a config file (e.g., `.aicrrc` or env vars):
用户应能通过配置文件或环境变量进行配置：

- **Target Branch**: Which branches to review (e.g., only `main` or `develop`). / 目标分支。
- **Review Language**: Output language (Chinese/English). / 审查语言。
- **Prompt Customization**: Ability to tweak the system prompt for the AI. / Prompt 自定义。
- **Severity Level**: Filter out "nitpick" comments if desired. / 严重等级过滤。

## 4. Security Requirements / 安全需求

- **Secrets Management**: API keys (OpenAI, GitHub Token) must never be logged.
- **密钥管理**: API Key 绝不能被记录日志。
- **Code Privacy**: Code snippets sent to AI must be minimized to the diff context.
- **代码隐私**: 发送给 AI 的代码片段必须最小化为 Diff 上下文。

## 5. User Stories / 用户故事

- **Developer**: I want my PR to get immediate feedback on common mistakes so I can fix them before assigning a human reviewer.
- **开发者**: 我希望我的 PR 能立即获得反馈，以便在指派人工审查前修复常见错误。
- **Team Lead**: I want to enforce coding standards automatically across the repo.
- **团队 Leader**: 我希望在仓库中自动强制执行编码规范。
- **DevOps Engineer**: I want to easily deploy this as a Docker container or a GitHub Action.
- **DevOps 工程师**: 我希望能轻松将其部署为 Docker 容器或 GitHub Action。
