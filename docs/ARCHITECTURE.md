# Technical Architecture & Design

[🇨🇳 中文](./ARCHITECTURE.zh-CN.md) | **🇺🇸 English**

## 1. Technology Stack

- **Runtime**: Node.js (v18+)
- **Language**: TypeScript (Strict typing for maintainability)
- **AI Model**: OpenAI API (GPT-4o/GPT-3.5) / Anthropic / Compatible APIs
- **Git Integration**:
  - `octokit` (GitHub)
  - `@gitbeaker/rest` (GitLab)
  - `gitea-js` or raw fetch (Gitea)
- **Server**: Koa or Fastify (Express currently used)
- **Utilities**: `zod` (Validation), `dotenv` (Config)

## 2. System Architecture

The project follows a **Clean Architecture / Hexagonal Architecture** to decouple the core logic from specific platforms.

### 2.1 High-Level Modules

```mermaid
graph TD
    User[User / Git Platform] -->|Webhook/CLI| Entry[Entry Point (Server/CLI)]
    Entry --> Core[Core Logic]

    subgraph Core Logic
        Orchestrator[Review Orchestrator]
        Queue[Concurrency Queue]
        DiffParser[Diff Parser]
        PromptEng[Prompt Engineer]
    end

    Core -->|Schedule| Queue
    Queue -->|Execute| IAI

    Core -->|Use| IPlatform[Interface: IPlatformProvider]
    Core -->|Use| IAI[Interface: IAIProvider]

    IPlatform -->|Implement| GitHub[GitHub Adapter]
    IPlatform -->|Implement| GitLab[GitLab Adapter]
    IPlatform -->|Implement| Gitea[Gitea Adapter]

    IAI -->|Implement| OpenAI[OpenAI Adapter]
    IAI -->|Implement| OtherAI[DeepSeek/Anthropic Adapter]
```

## 3. Module Details

### 3.1 Platform Adapter Layer (`src/platforms/`)

All platforms must implement the `IPlatformProvider` interface:

```typescript
interface IPlatformProvider {
  // Parse incoming webhook payload to a standard Event object (Async)
  parseEvent(request: Request): Promise<ReviewEvent | null>;

  // Fetch the file diffs for a specific PR
  getChanges(eventId: string): Promise<FileChange[]>;

  // Post comments back to the platform
  postComment(eventId: string, comment: ReviewComment): Promise<void>;
}
```

### 3.2 Core Logic (`src/core/`)

- **Review Orchestrator**: Controls the flow: Receive Event -> Fetch Diff -> Split into Chunks -> Send to AI -> Parse Response -> Post Comments.
- **Diff Parser**: Parses unified diff format into manageable chunks with line numbers.

### 3.3 AI Service (`src/ai/`)

- Handles token counting (to avoid context window overflow).
- Manages prompts (System prompts, few-shot examples).
- Parses structured JSON output from LLM to map comments to line numbers.

## 4. Data Flow

1.  **Trigger**: Webhook received from Gitea/GitHub/GitLab.
2.  **Normalization**: Request converted to internal `ReviewContext` (repo, pr_id, commit_sha).
3.  **Fetch**: System calls `platform.getDiff()`.
4.  **Parsing**: Raw diff string parsed into `FileDiff` objects (filename, chunks, line_map).
5.  **Filtering**: Files matching `.gitignore` or specific extensions are dropped.
6.  **Analysis**:
    - Loop through files.
    - Construct prompt: "Review this diff for file X".
    - Call AI Provider.
7.  **Response Handling**: AI returns JSON array of issues `[{ line: 10, content: "Fix X" }]`.
8.  **Posting**: System calls `platform.postReview()` to create inline comments.

## 5. Directory Structure

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
