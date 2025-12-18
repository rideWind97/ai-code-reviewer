# 使用指南 (Usage Guide)

本文档详细说明了如何配置、部署和使用 AI Code Reviewer。

## 1. 环境变量配置 (.env)

在项目根目录创建 `.env` 文件：

```ini
# --- 服务配置 ---
PORT=3000
NODE_ENV=production

# --- AI 服务配置 ---
AI_PROVIDER=openai
# OpenAI API Key (必填)
AI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxx
# 模型名称 (推荐 gpt-4o, gpt-3.5-turbo)
AI_MODEL=gpt-4o
# 可选：自定义 API 地址 (用于代理或兼容模型)
AI_BASE_URL=https://api.openai.com/v1

# --- Git 平台配置 (按需填写) ---

# GitHub
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxx    # 个人访问令牌 (PAT)
GITHUB_WEBHOOK_SECRET=your_secret    # Webhook 签名密钥

# GitLab
GITLAB_HOST=https://gitlab.com       # 自托管请修改域名
GITLAB_TOKEN=glpat-xxxxxxxxxxxx      # Access Token

# Gitea
GITEA_HOST=https://gitea.example.com
GITEA_TOKEN=xxxxxxxxxxxxxxxxxxxx

# --- 审查规则 ---
# 目标分支 (逗号分隔)
TARGET_BRANCHES=main,master,develop
# 忽略的文件类型或路径
IGNORE_PATTERNS=package-lock.json,dist,node_modules,yarn.lock
```

## 2. 部署模式

### 模式 A: Webhook Server (推荐)

作为独立服务运行，实时监听代码平台的 Webhook 事件。

1.  **启动服务**:

    ```bash
    docker-compose up -d
    ```

    服务默认运行在 `3000` 端口。

2.  **配置 Webhook**:
    - **GitHub**: Repository Settings -> Webhooks -> Add webhook
      - Payload URL: `http://your-server-ip:3000/webhook`
      - Content type: `application/json`
      - Secret: (对应 .env 中的 GITHUB_WEBHOOK_SECRET)
      - Events: 选择 `Pull requests`
    - **GitLab**: Settings -> Webhooks
      - URL: `http://your-server-ip:3000/webhook`
      - Trigger: `Merge request events`
    - **Gitea**: 仓库设置 -> Web钩子
      - 目标 URL: `http://your-server-ip:3000/webhook`
      - 事件: `Pull Request`

### 模式 B: CLI / CI 集成

作为 CI/CD 流水线的一个步骤运行，无需维护长期服务。

#### 1. GitHub Actions

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
        run: npm install -g ai-code-reviewer # 或者 git clone 后 npm install

      - name: Run Review
        # 假设你已经 clone 了本项目代码或者发布了包
        # 这里演示源码运行方式
        run: npx ts-node src/cli/index.ts run --platform github
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          # GITHUB_EVENT_PATH 会由 Actions 自动注入
```

#### 2. GitLab CI

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

#### 3. 手动运行 (CLI)

你也可以在本地手动触发审查：

```bash
# 审查 GitHub PR #123
npx ts-node src/cli/index.ts run \
  --platform github \
  --repo owner/repo \
  --pr 123 \
  --sha <head_commit_sha>
```

## 3. 开发指南

### 本地调试 CLI

```bash
# 编译
npm run build

# 链接命令 (可选)
npm link

# 运行
aicr run --help
```

### 常见问题

1.  **GitLab 评论失败**:
    - 确保 `.env` 中正确配置了 `GITLAB_HOST` 和 `GITLAB_TOKEN`。
    - GitLab 内联评论需要准确的 `base_sha`, `head_sha` 和 `start_sha`。CLI 模式下请务必传递 `--base-sha`。

2.  **AI 响应慢**:
    - 检查 Diff 大小。建议在 `.env` 中配置 `IGNORE_PATTERNS` 过滤掉大文件。
    - OpenAI API 偶尔会有延迟，可尝试更换兼容的国内 LLM 服务。
