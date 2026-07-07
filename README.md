# AICR

AICR 是一个用 Go 编写的 GitHub PR 自动 Code Review 服务骨架。当前 MVP 只做“审查并评论”：接收 GitHub webhook，拉取 PR diff，拆分 review unit，调用模型审查，聚合 findings，最后通过 provider adapter 发布 summary、inline comment 和 status。

## 当前 MVP 范围

- GitHub pull request webhook 验签和幂等入队。
- GitLab merge request webhook token 校验和幂等入队。
- Provider-neutral 的 patch/comment/status 接口，以及 GitHub/GitLab adapter。
- Review run 状态机和进程内 runner。
- Review unit 拆分、Context Pack 构建、Model Gateway、finding 聚合、发布辅助函数和 secret scanning。
- Large/Huge PR/MR 分级、diff hunk 拆分、模块摘要和动态 inline 评论预算。
- 持久化队列接口、`database/sql` PostgreSQL 队列适配器、worker 常驻消费编排、PostgreSQL store 骨架。
- Project Analyzer、仓库级 `.ai-review/SKILL.md` 解析、`/ai-review` comment command 解析。
- GitHub App installation token 缓存骨架、模型 JSON repair、发布前安全扫描包装器。
- `migrations/` 下的 PostgreSQL 初始 schema。

repo cache 本地 diff fallback、真实 GitHub App JWT 换 token、多模型 fallback 和完整治理审计仍是后续增强。

## 目录结构

```text
.
├── cmd/
│   ├── aicr-server/          # HTTP 服务入口：health check、GitHub webhook
│   └── aicr-worker/          # 后台 worker 入口：后续消费队列并执行 review run
├── internal/
│   ├── config/               # 读取环境变量，组装服务配置
│   ├── httpserver/           # HTTP 路由，把具体 handler 挂到路径上
│   ├── model/                # 模型网关接口和 OpenAI-compatible client
│   ├── commentcommand/       # 解析 /ai-review、/ai-review full、/ai-review file ...
│   ├── provider/             # Provider-neutral 类型和接口
│   │   ├── github/           # GitHub webhook 和 GitHub REST API adapter
│   │   ├── githubauth/       # GitHub App installation token 缓存骨架
│   │   └── gitlab/           # GitLab MR webhook 和 GitLab REST API adapter
│   ├── project/
│   │   ├── analyzer/         # 生成项目分析摘要
│   │   └── skill/            # 解析和合并仓库级 review skill
│   ├── queue/                # 队列接口；支持内存队列和 database/sql PostgreSQL 队列
│   ├── review/
│   │   ├── agent/            # 把 Context Pack 转成 prompt，调用模型审查
│   │   ├── aggregate/        # 对 findings 做证据等级过滤、去重和评论预算控制
│   │   ├── chunker/          # 把 PatchSet 拆成 ReviewUnit
│   │   ├── context/          # 为每个 ReviewUnit 构建 Context Pack
│   │   ├── orchestrator/     # Review run 状态机和主流程 runner
│   │   ├── publish/          # 构造 summary comment 和 inline comments
│   │   ├── scale/            # 判断 PR/MR 是 small、large 还是 huge，并给出动态策略
│   │   └── summary/          # 按模块聚合文件数和增删行，服务 large/huge summary
│   ├── security/             # 发布前安全扫描包装器
│   ├── secret/               # 发布前 secret scanning 的轻量规则
│   ├── store/                # 持久化接口和 PostgreSQL store 骨架
│   └── worker/               # 从队列消费 review job，按 provider 路由到 runner
├── migrations/               # 数据库迁移 SQL
├── test/integration/         # 端到端 mock 测试
├── docker-compose.yml        # 本地 PostgreSQL
└── Makefile                  # 常用开发命令
```

建议 Go 新手从 `cmd/aicr-server/main.go` 看起，然后按 `webhook -> queue -> orchestrator -> chunker/context -> agent/model -> aggregate/publish` 的顺序阅读。

## Large/Huge PR 策略

- `internal/review/scale` 会按变更文件数和 diff 增删行数把 PR/MR 分为 `small`、`large`、`huge`。
- `small` 使用默认策略：按文件拆 review unit，最多发布 5 条 inline comment。
- `large` 会打开 hunk 拆分，把大文件中的多个 `@@` diff 片段拆成更小 review unit，并把 inline 预算降到 3 条。
- `huge` 默认不发布 inline comment，只保留 summary 和风险热点，避免在超大 PR/MR 中刷屏。
- `internal/review/summary` 可以按顶层目录生成模块摘要，帮助 summary comment 快速说明哪些模块变化最大。

## 配置

环境变量：

- `AICR_ADDR`：HTTP 监听地址，默认 `:8080`。
- `AICR_DATABASE_URL`：PostgreSQL 连接串；配置后 server 会把 webhook job 写入 PostgreSQL，worker 会从 PostgreSQL 消费。
- `AICR_GITHUB_WEBHOOK_SECRET`：GitHub webhook secret，用于校验请求确实来自 GitHub。
- `AICR_GITHUB_TOKEN`：当前 GitHub adapter 骨架使用的 GitHub API token。
- `AICR_GITLAB_BASE_URL`：GitLab base URL；为空时默认使用 GitLab.com。
- `AICR_GITLAB_TOKEN`：GitLab API token，可以是 OAuth、project、group 或 bot token。
- `AICR_GITLAB_SECRET_TOKEN`：GitLab webhook Secret Token，用于校验 `X-Gitlab-Token`。
- `AICR_MODEL_BASE_URL`：OpenAI-compatible API base URL。
- `AICR_MODEL_NAME`：模型名称。
- `AICR_MODEL_API_KEY`：模型 API key。

## 开发

启动 Postgres：

```bash
docker compose up -d postgres
```

运行测试：

```bash
make test
```

运行 HTTP 服务：

```bash
make run-server
```

运行 worker：

```bash
make run-worker
```

`aicr-worker` 需要 `AICR_DATABASE_URL`。它会持续轮询 `review_jobs` 表，领取 `queued` 状态任务，并按 `provider` 字段路由到 GitHub 或 GitLab runner。

## 架构说明

实现遵循 `docs/ai-coding-agent-architecture.md`，采用模块化单体。`internal/review` 只关心 review pipeline，不直接依赖 GitHub 或 GitLab；`internal/provider/github` 和 `internal/provider/gitlab` 分别负责各自 webhook 与 REST API 细节。这样新增 provider 时，可以复用 review pipeline，只新增 provider adapter。
