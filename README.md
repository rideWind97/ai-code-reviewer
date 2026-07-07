# AICR

AICR 是一个用 Go 编写的 GitHub PR 自动 Code Review 服务骨架。当前 MVP 只做“审查并评论”：接收 GitHub webhook，拉取 PR diff，拆分 review unit，调用模型审查，聚合 findings，最后通过 provider adapter 发布 summary、inline comment 和 status。

## 当前 MVP 范围

- GitHub pull request webhook 验签和幂等入队。
- GitLab merge request webhook token 校验和幂等入队。
- Provider-neutral 的 patch/comment/status 接口，以及 GitHub/GitLab adapter。
- Review run 状态机和进程内 runner。
- Review unit 拆分、Context Pack 构建、Model Gateway、finding 聚合、发布辅助函数和 secret scanning。
- `migrations/` 下的 PostgreSQL 初始 schema。

持久化队列 worker、GitHub App installation token 换取、repo cache、超大 PR 摘要、项目级 review skill 都是后续阶段。

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
│   ├── provider/             # Provider-neutral 类型和接口
│   │   ├── github/           # GitHub webhook 和 GitHub REST API adapter
│   │   └── gitlab/           # GitLab MR webhook 和 GitLab REST API adapter
│   ├── queue/                # 队列接口；当前用内存队列实现 MVP 幂等入队
│   ├── review/
│   │   ├── agent/            # 把 Context Pack 转成 prompt，调用模型审查
│   │   ├── aggregate/        # 对 findings 做证据等级过滤、去重和评论预算控制
│   │   ├── chunker/          # 把 PatchSet 拆成 ReviewUnit
│   │   ├── context/          # 为每个 ReviewUnit 构建 Context Pack
│   │   ├── orchestrator/     # Review run 状态机和主流程 runner
│   │   └── publish/          # 构造 summary comment 和 inline comments
│   ├── secret/               # 发布前 secret scanning 的轻量规则
│   └── store/                # 持久化接口，后续接 PostgreSQL repository
├── migrations/               # 数据库迁移 SQL
├── test/integration/         # 端到端 mock 测试
├── docker-compose.yml        # 本地 PostgreSQL
└── Makefile                  # 常用开发命令
```

建议 Go 新手从 `cmd/aicr-server/main.go` 看起，然后按 `webhook -> queue -> orchestrator -> chunker/context -> agent/model -> aggregate/publish` 的顺序阅读。

## 配置

环境变量：

- `AICR_ADDR`：HTTP 监听地址，默认 `:8080`。
- `AICR_DATABASE_URL`：PostgreSQL 连接串。
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

运行 worker 占位入口：

```bash
make run-worker
```

## 架构说明

实现遵循 `docs/ai-coding-agent-architecture.md`，采用模块化单体。`internal/review` 只关心 review pipeline，不直接依赖 GitHub 或 GitLab；`internal/provider/github` 和 `internal/provider/gitlab` 分别负责各自 webhook 与 REST API 细节。这样新增 provider 时，可以复用 review pipeline，只新增 provider adapter。
