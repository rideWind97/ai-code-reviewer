# Technical Architecture & Design / 技术架构设计

## 1. Technology Stack / 技术栈

- **Runtime**: Node.js (v18+)
- **Language**: TypeScript (Strict typing for maintainability / 强类型)
- **AI Model**: OpenAI API (GPT-4o/GPT-3.5) / Anthropic / Compatible APIs
- **Git Integration**:
  - `octokit` (GitHub)
  - `@gitbeaker/rest` (GitLab)
  - `gitea-js` or raw fetch (Gitea)
- **Server**: Koa or Fastify (Express currently used / 当前使用 Express)
- **Utilities**: `zod` (Validation), `dotenv` (Config)

## 2. System Architecture / 系统架构

The project follows a **Clean Architecture / Hexagonal Architecture** to decouple the core logic from specific platforms.
项目采用 **整洁架构 / 六边形架构**，将核心逻辑与具体平台解耦。

### 2.1 High-Level Modules / 高层模块

```mermaid
graph TD
    User[User / Git Platform] -->|Webhook/CLI| Entry[Entry Point (Server/CLI)]
    Entry --> Core[Core Logic]

    subgraph Core Logic
        Orchestrator[Review Orchestrator]
        DiffParser[Diff Parser]
        PromptEng[Prompt Engineer]
    end

    Core -->|Use| IPlatform[Interface: IPlatformProvider]
    Core -->|Use| IAI[Interface: IAIProvider]

    IPlatform -->|Implement| GitHub[GitHub Adapter]
    IPlatform -->|Implement| GitLab[GitLab Adapter]
    IPlatform -->|Implement| Gitea[Gitea Adapter]

    IAI -->|Implement| OpenAI[OpenAI Adapter]
    IAI -->|Implement| OtherAI[DeepSeek/Anthropic Adapter]
```

## 3. Module Details / 模块详情

### 3.1 Platform Adapter Layer / 平台适配层 (`src/platforms/`)

All platforms must implement the `IPlatformProvider` interface:
所有平台必须实现 `IPlatformProvider` 接口：

```typescript
interface IPlatformProvider {
  // Parse incoming webhook payload to a standard Event object
  // 将 Webhook 负载解析为标准事件对象
  parseEvent(request: Request): Promise<ReviewEvent | null>;

  // Fetch the file diffs for a specific PR
  // 获取特定 PR 的文件差异
  getChanges(eventId: string): Promise<FileChange[]>;

  // Post comments back to the platform
  // 回写评论到平台
  postComment(eventId: string, comment: ReviewComment): Promise<void>;
}
```

### 3.2 Core Logic / 核心逻辑 (`src/core/`)

- **Review Orchestrator**: Controls the flow: Receive Event -> Fetch Diff -> Split into Chunks -> Send to AI -> Parse Response -> Post Comments.
- **Review Orchestrator**: 流程控制：接收事件 -> 获取 Diff -> 分块 -> AI 分析 -> 解析响应 -> 发布评论。
- **Diff Parser**: Parses unified diff format into manageable chunks with line numbers.
- **Diff Parser**: 将统一 Diff 格式解析为带有行号的代码块。

### 3.3 AI Service / AI 服务 (`src/ai/`)

- Handles token counting (to avoid context window overflow).
- 处理 Token 计数（避免上下文溢出）。
- Manages prompts (System prompts, few-shot examples).
- 管理 Prompts。
- Parses structured JSON output from LLM to map comments to line numbers.
- 解析 LLM 的 JSON 输出并映射行号。

## 4. Data Flow / 数据流

1.  **Trigger**: Webhook received from Gitea/GitHub/GitLab. (收到 Webhook)
2.  **Normalization**: Request converted to internal `ReviewContext` (repo, pr_id, commit_sha). (归一化请求)
3.  **Fetch**: System calls `platform.getDiff()`. (获取 Diff)
4.  **Parsing**: Raw diff string parsed into `FileDiff` objects (filename, chunks, line_map). (解析 Diff)
5.  **Filtering**: Files matching `.gitignore` or specific extensions are dropped. (过滤文件)
6.  **Analysis**:
    - Loop through files. (遍历文件)
    - Construct prompt: "Review this diff for file X". (构造 Prompt)
    - Call AI Provider. (调用 AI)
7.  **Response Handling**: AI returns JSON array of issues `[{ line: 10, content: "Fix X" }]`. (处理响应)
8.  **Posting**: System calls `platform.postReview()` to create inline comments. (发布评论)

## 5. Directory Structure / 目录结构

```
src/
├── ai/                 # AI Provider implementations
├── config/             # Configuration loading
├── core/               # Business logic (Orchestrator, Parser)
├── interfaces/         # TypeScript interfaces
├── platforms/          # GitHub, GitLab, Gitea adapters
├── server/             # Webhook server entry point
├── cli/                # CLI entry point
└── utils/              # Helper functions
```
