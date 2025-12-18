# 技术架构设计

**🇨🇳 中文** | [🇺🇸 English](./ARCHITECTURE.md)

## 1. 技术栈

- **运行环境**: Node.js (v18+)
- **语言**: TypeScript (强类型维护)
- **AI 模型**: OpenAI API (GPT-4o/GPT-3.5) / Anthropic / 兼容接口
- **Git 集成**:
  - `octokit` (GitHub)
  - `@gitbeaker/rest` (GitLab)
  - `gitea-js` 或原生 fetch (Gitea)
- **服务器**: Koa 或 Fastify (当前使用 Express)
- **工具库**: `zod` (验证), `dotenv` (配置)

## 2. 系统架构

项目采用 **整洁架构 / 六边形架构**，将核心逻辑与具体平台解耦。

### 2.1 高层模块

```mermaid
graph TD
    User[用户 / Git 平台] -->|Webhook/CLI| Entry[入口 (Server/CLI)]
    Entry --> Core[核心逻辑]

    subgraph Core Logic
        Orchestrator[审查编排器]
        DiffParser[Diff 解析器]
        PromptEng[Prompt 引擎]
    end

    Core -->|使用| IPlatform[接口: IPlatformProvider]
    Core -->|使用| IAI[接口: IAIProvider]

    IPlatform -->|实现| GitHub[GitHub 适配器]
    IPlatform -->|实现| GitLab[GitLab 适配器]
    IPlatform -->|实现| Gitea[Gitea 适配器]

    IAI -->|实现| OpenAI[OpenAI 适配器]
    IAI -->|实现| OtherAI[DeepSeek/Anthropic 适配器]
```

## 3. 模块详情

### 3.1 平台适配层 (`src/platforms/`)

所有平台必须实现 `IPlatformProvider` 接口：

```typescript
interface IPlatformProvider {
  // 将 Webhook 负载解析为标准事件对象
  parseEvent(request: Request): Promise<ReviewEvent | null>;

  // 获取特定 PR 的文件差异
  getChanges(eventId: string): Promise<FileChange[]>;

  // 回写评论到平台
  postComment(eventId: string, comment: ReviewComment): Promise<void>;
}
```

### 3.2 核心逻辑 (`src/core/`)

- **Review Orchestrator (编排器)**: 流程控制：接收事件 -> 获取 Diff -> 分块 -> AI 分析 -> 解析响应 -> 发布评论。
- **Diff Parser (Diff 解析)**: 将统一 Diff 格式解析为带有行号的代码块。

### 3.3 AI 服务 (`src/ai/`)

- 处理 Token 计数（避免上下文溢出）。
- 管理 Prompts。
- 解析 LLM 的 JSON 输出并映射行号。

## 4. 数据流

1.  **触发**: 收到 Gitea/GitHub/GitLab 的 Webhook。
2.  **归一化**: 请求转换为内部 `ReviewContext` (repo, pr_id, commit_sha)。
3.  **获取**: 系统调用 `platform.getDiff()`。
4.  **解析**: 原始 Diff 字符串解析为 `FileDiff` 对象 (文件名, 块, 行号映射)。
5.  **过滤**: 忽略 `.gitignore` 匹配的文件或特定扩展名。
6.  **分析**:
    - 遍历文件。
    - 构造 Prompt: "审查文件 X 的这个 Diff"。
    - 调用 AI Provider。
7.  **响应处理**: AI 返回问题的 JSON 数组 `[{ line: 10, content: "修复 X" }]`。
8.  **发布**: 系统调用 `platform.postReview()` 创建内联评论。

## 5. 目录结构

```
src/
├── ai/                 # AI Provider 实现
├── config/             # 配置加载
├── core/               # 业务逻辑 (Orchestrator, Parser)
├── interfaces/         # TypeScript 接口定义
├── platforms/          # GitHub, GitLab, Gitea 适配器
├── server/             # Webhook 服务器入口
├── cli/                # CLI 入口
└── utils/              # 辅助函数
```
