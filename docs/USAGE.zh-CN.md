# 使用指南

**🇨🇳 中文** | [🇺🇸 English](./USAGE.md)

本文档提供了配置、部署和使用 AI Code Reviewer 的详细指南。

## 1. 配置详情

在项目根目录创建 `.env` 文件。以下是所有支持的环境变量的详细说明。

### 🤖 AI 服务配置

| 变量名 | 是否必填 | 描述 | 示例 |
| :--- | :--- | :--- | :--- |
| `AI_PROVIDER` | **是** | AI 服务提供商。 | `openai`, `deepseek` |
| `AI_API_KEY` | **是** | 你的 API Key。 | `sk-proj-...` |
| `AI_MODEL` | 否 | 模型名称。默认为 `gpt-4o`。 | `gpt-4o`, `deepseek-chat` |
| `AI_BASE_URL` | 否 | 自定义 API 地址 (用于代理或兼容模型)。 | `https://api.deepseek.com` |

### 🛠 系统配置

| 变量名 | 默认值 | 描述 |
| :--- | :--- | :--- |
| `PORT` | `3000` | Webhook 服务监听端口。 |
| `NODE_ENV` | `production` | 设置为 `development` 可开启调试日志和请求转储。 |
| `CONCURRENCY_LIMIT` | `5` | 最大并发处理文件数，防止 API 速率限制。 |

### 📝 审查规则

| 变量名 | 默认值 | 描述 |
| :--- | :--- | :--- |
| `TARGET_BRANCHES` | `main,master` | 仅审查合并到这些分支的 PR。 |
| `IGNORE_PATTERNS` | `lock,dist` | 忽略的文件名关键字 (逗号分隔)。 |
| `REVIEW_LANGUAGE` | `zh-CN` | AI 评论使用的语言 (`en`, `zh-CN`, `ja` 等)。 |
| `SKIP_DRAFT_PR` | `true` | 是否跳过 Draft/WIP 状态的 PR。 |

### 🔌 平台配置

仅需配置你使用的平台。

**GitHub**
```bash
GITHUB_TOKEN=ghp_...          # 个人访问令牌 (PAT)
GITHUB_WEBHOOK_SECRET=...     # GitHub Webhook 设置中的 Secret
```

**GitLab**
```bash
GITLAB_HOST=https://gitlab.com # 私有部署请填写你的域名
GITLAB_TOKEN=glpat-...         # Access Token
```

**Gitea**
```bash
GITEA_HOST=https://gitea.com
GITEA_TOKEN=...
```

---

## 2. 部署模式

### 模式 A: Webhook Server (推荐)

适合团队使用。作为常驻服务 (Docker) 运行，自动监听代码平台的事件。

1.  **启动服务**:
    ```bash
    docker-compose up -d
    ```

2.  **配置 Webhook**:
    前往仓库设置页面添加 Webhook:
    *   **Payload URL**: `http://<your-server-ip>:4000/webhook`
    *   **Content Type**: `application/json`
    *   **Secret**: 必须匹配 `.env` 中的 `GITHUB_WEBHOOK_SECRET`。
    *   **触发事件**:
        *   GitHub: `Pull requests`, `Issue comments`
        *   GitLab: `Merge request events`, `Note events`
        *   Gitea: `Pull Request`, `Issue Comment`

### 模式 B: CLI / CI 集成

适合按需运行，无需维护服务器。

**示例: GitHub Actions**

添加文件 `.github/workflows/review.yml`:

```yaml
name: AI Code Review
on: [pull_request]
permissions:
  pull-requests: write
  contents: read

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install Tool
        run: npm install -g ai-code-reviewer
      - name: Run Review
        run: aicr run --platform github
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

---

## 3. ChatOps (交互式指令)

你可以在 Pull Request 的评论区直接与机器人交互。

### 指令: `/review`

**场景**: 你刚刚推送了修复代码，希望 AI 立即重新审查，而不需要等待新的 Commit 或推送空 Commit。

1.  进入 PR 的 **Conversation (对话)** 标签页。
2.  发表评论: `/review`。
3.  机器人将获取最新的 Diff 并发布新的审查意见。

*(注意: 需要在 Webhook 设置中开启 `Issue comments` 或 `Note events` 权限)*

---

## 4. 自定义 AI 角色 (Prompt)

你可以通过在项目根目录创建 `prompt.md` 文件来定制 AI 的审查风格（例如：更严格、专注于安全、或使用特定语气）。

**示例 `prompt.md`**:

```markdown
你是一位资深的前端工程师。
请审查代码中的以下问题：
1. 潜在的运行时错误。
2. 变量命名拼写错误。
3. 性能问题 (如不必要的渲染)。

风格: 简洁专业，不要吝啬批评。
语言: 中文。
```

系统会自动加载该文件并用于下一次审查。

---

## 5. 开发与排错

### 本地调试 CLI
无需全局安装即可在本地测试 CLI：

```bash
# 1. 编译
npm run build

# 2. 运行 CLI
# 将 <sha> 替换为仓库中真实的 commit SHA
npx ts-node src/cli/index.ts run \
  --platform github \
  --repo owner/repo \
  --pr 1 \
  --sha <head_sha>
```

### 常见问题

**Q: 机器人审查了已删除的文件。**
A: 请确保 `IGNORE_PATTERNS` 包含了 `lock` 或特定的构建目录。虽然机器人会自动检查文件状态，但在某些 Diff 模式下上下文可能不明确。

**Q: GitLab 评论没有显示。**
A: GitLab API 需要精确的 `old_line` 和 `new_line` 引用。如果 Diff 已经过时，评论可能会失败。尝试使用 `/review` 重新触发。

**Q: Docker 中出现 `EPERM` 或网络错误。**
A: 检查 Docker 容器是否有权限访问外部网络（尤其是 Git 平台是私有部署时，需确保 DNS 配置正确）。
