# Usage Guide / 使用指南

This document details how to configure, deploy, and use the AI Code Reviewer.
本文档详细说明了如何配置、部署和使用 AI Code Reviewer。

## 1. Environment Configuration / 环境变量配置 (.env)

Create a `.env` file in the project root:
在项目根目录创建 `.env` 文件：

```ini
# --- Service Configuration / 服务配置 ---
PORT=3000
NODE_ENV=production

# --- AI Service Configuration / AI 服务配置 ---
AI_PROVIDER=openai
# OpenAI API Key (Required / 必填)
AI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxx
# Model Name (Recommended: gpt-4o, gpt-3.5-turbo / 推荐 gpt-4o)
AI_MODEL=gpt-4o
# Optional: Custom API URL (For proxies or compatible models / 可选：自定义 API 地址)
AI_BASE_URL=https://api.openai.com/v1

# --- Git Platform Configuration (Fill as needed) / Git 平台配置 (按需填写) ---

# GitHub
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxx    # Personal Access Token (PAT)
GITHUB_WEBHOOK_SECRET=your_secret    # Webhook Secret

# GitLab
GITLAB_HOST=https://gitlab.com       # Change if self-hosted / 自托管请修改域名
GITLAB_TOKEN=glpat-xxxxxxxxxxxx      # Access Token

# Gitea
GITEA_HOST=https://gitea.example.com
GITEA_TOKEN=xxxxxxxxxxxxxxxxxxxx

# --- Review Rules / 审查规则 ---
# Target Branches (Comma separated) / 目标分支 (逗号分隔)
TARGET_BRANCHES=main,master,develop
# Ignored file patterns / 忽略的文件类型或路径
IGNORE_PATTERNS=package-lock.json,dist,node_modules,yarn.lock
```

## 2. Deployment Modes / 部署模式

### Mode A: Webhook Server (Recommended) / 模式 A: Webhook Server (推荐)

Run as a standalone service, listening for Webhook events in real-time.
作为独立服务运行，实时监听代码平台的 Webhook 事件。

1.  **Start Service / 启动服务**:

    ```bash
    docker-compose up -d
    ```

    Service runs on port `3000` by default.
    服务默认运行在 `3000` 端口。

2.  **Configure Webhook / 配置 Webhook**:
    - **GitHub**: Repository Settings -> Webhooks -> Add webhook
      - Payload URL: `http://your-server-ip:3000/webhook`
      - Content type: `application/json`
      - Secret: (Match .env `GITHUB_WEBHOOK_SECRET` / 对应 .env 配置)
      - Events: Select `Pull requests`
    - **GitLab**: Settings -> Webhooks
      - URL: `http://your-server-ip:3000/webhook`
      - Trigger: `Merge request events`
    - **Gitea**: Settings -> Webhooks / 仓库设置 -> Web钩子
      - Target URL / 目标 URL: `http://your-server-ip:3000/webhook`
      - Events / 事件: `Pull Request`

### Mode B: CLI / CI Integration / 模式 B: CLI / CI 集成

Run as a step in your CI/CD pipeline.
作为 CI/CD 流水线的一个步骤运行，无需维护长期服务。

#### 1. GitHub Actions

Add to your `.github/workflows/ai-review.yml`:
在你的仓库 `.github/workflows/ai-review.yml` 中添加：

```yaml
name: AI Code Review
on: [pull_request]

permissions:
  contents: read
  pull-requests: write

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Node
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install dependencies
        run: npm install -g ai-code-reviewer # Or clone and npm install

      - name: Run Review
        # Assuming source code checkout
        # 假设你已经 clone 了本项目代码
        run: npx ts-node src/cli/index.ts run --platform github
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          # GITHUB_EVENT_PATH is injected by Actions automatically
```

#### 2. GitLab CI

In `.gitlab-ci.yml`:
在 `.gitlab-ci.yml` 中：

```yaml
ai_review:
  image: node:18
  stage: test
  script:
    - npm install
    - npx ts-node src/cli/index.ts run --platform gitlab --repo $CI_PROJECT_PATH --pr $CI_MERGE_REQUEST_IID --sha $CI_COMMIT_SHA --base-sha $CI_MERGE_REQUEST_DIFF_BASE_SHA
  only:
    - merge_requests
  variables:
    GITLAB_TOKEN: $GITLAB_ACCESS_TOKEN
    AI_API_KEY: $OPENAI_API_KEY
```

#### 3. Manual Execution (CLI) / 手动运行 (CLI)

You can trigger a review locally:
你也可以在本地手动触发审查：

```bash
# Review GitHub PR #123
npx ts-node src/cli/index.ts run \
  --platform github \
  --repo owner/repo \
  --pr 123 \
  --sha <head_commit_sha>
```

## 3. Development / 开发指南

### Debug CLI Locally / 本地调试 CLI

```bash
# Build / 编译
npm run build

# Link Command (Optional) / 链接命令 (可选)
npm link

# Run / 运行
aicr run --help
```

### FAQ / 常见问题

1.  **GitLab Comment Failed / GitLab 评论失败**:
    - Ensure `GITLAB_HOST` and `GITLAB_TOKEN` are correct.
    - 确保 `.env` 中正确配置了 `GITLAB_HOST` 和 `GITLAB_TOKEN`。
    - GitLab inline comments require exact `base_sha`, `head_sha`, and `start_sha`. Provide `--base-sha` in CLI mode.
    - GitLab 内联评论需要准确的 `base_sha`, `head_sha` 和 `start_sha`。CLI 模式下请务必传递 `--base-sha`。

2.  **Slow AI Response / AI 响应慢**:
    - Check Diff size. Use `IGNORE_PATTERNS` to filter large files.
    - 检查 Diff 大小。建议在 `.env` 中配置 `IGNORE_PATTERNS` 过滤掉大文件。
    - OpenAI API latency. Try compatible local models or proxies.
    - OpenAI API 偶尔会有延迟，可尝试更换兼容的国内 LLM 服务。
