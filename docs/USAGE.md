# Usage Guide

[🇨🇳 中文](./USAGE.zh-CN.md) | **🇺🇸 English**

This document provides a comprehensive guide on configuring, deploying, and using the AI Code Reviewer.

## 1. Configuration Reference

Create a `.env` file in the project root. Below is a detailed explanation of all supported environment variables.

### 🤖 AI Service Configuration

| Variable | Required | Description | Example |
| :--- | :--- | :--- | :--- |
| `AI_PROVIDER` | **Yes** | The AI provider to use. | `openai`, `deepseek` |
| `AI_API_KEY` | **Yes** | Your API Key. | `sk-proj-...` |
| `AI_MODEL` | No | The model ID. Defaults to `gpt-4o`. | `gpt-4o`, `gpt-3.5-turbo`, `deepseek-chat` |
| `AI_BASE_URL` | No | Custom API endpoint for proxies or other providers. | `https://api.deepseek.com` |

### 🛠 System Configuration

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `3000` | The port the webhook server listens on. |
| `NODE_ENV` | `production` | Set to `development` to enable debug logs and dump requests. |
| `CONCURRENCY_LIMIT` | `5` | Max number of files processed concurrently to avoid rate limits. |

### 📝 Review Rules

| Variable | Default | Description |
| :--- | :--- | :--- |
| `TARGET_BRANCHES` | `main,master` | Only review PRs targeting these branches. |
| `IGNORE_PATTERNS` | `lock,dist` | Comma-separated list of keywords to ignore in filenames. |
| `REVIEW_LANGUAGE` | `zh-CN` | The language AI uses for comments (`en`, `zh-CN`, `ja`, etc.). |
| `SKIP_DRAFT_PR` | `true` | If `true`, the bot ignores Draft/WIP PRs. |

### 🔌 Platform Configuration

You only need to configure the platform you are using.

**GitHub**
```bash
GITHUB_TOKEN=ghp_...          # Personal Access Token (PAT) with repo scope
GITHUB_WEBHOOK_SECRET=...     # Secret set in GitHub Webhook settings
```

**GitLab**
```bash
GITLAB_HOST=https://gitlab.com # Or your self-hosted URL
GITLAB_TOKEN=glpat-...         # Access Token with api scope
```

**Gitea**
```bash
GITEA_HOST=https://gitea.com
GITEA_TOKEN=...
```

---

## 2. Deployment Modes

### Mode A: Webhook Server (Recommended)

Ideal for teams. Run it as a long-running service (Docker) that listens to events automatically.

1.  **Start the Service**:
    ```bash
    docker-compose up -d
    ```

2.  **Configure Webhook**:
    Go to your repository settings and add a Webhook:
    *   **Payload URL**: `http://<your-server-ip>:4000/webhook`
    *   **Content Type**: `application/json`
    *   **Secret**: Must match `GITHUB_WEBHOOK_SECRET` in `.env`.
    *   **Events**:
        *   GitHub: `Pull requests`, `Issue comments`
        *   GitLab: `Merge request events`, `Note events`
        *   Gitea: `Pull Request`, `Issue Comment`

### Mode B: CLI / CI Integration

Ideal for on-demand usage or keeping infrastructure stateless.

**Example: GitHub Actions**

Add `.github/workflows/review.yml`:

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

## 3. ChatOps (Interactive Commands)

You can interact with the bot directly in the Pull Request comment section.

### Command: `/review`

**Scenario**: You have pushed a fix and want the AI to re-evaluate the changes immediately without waiting for a new commit or pushing a dummy commit.

1.  Go to the **Conversation** tab of the PR.
2.  Post a comment: `/review`.
3.  The bot will fetch the latest Diff and post new reviews.

*(Note: Requires `Issue comments` or `Note events` webhook permission)*

---

## 4. Customizing the AI Persona

You can change how the AI reviews your code (e.g., be more strict, focus on security, or use a specific tone) by creating a `prompt.md` file in the root directory.

**Example `prompt.md`**:

```markdown
You are a Senior Typescript Engineer.
Review the code for:
1. Potential runtime errors.
2. Typos in variable names.
3. Performance issues.

Style: Be concise and professional. Do not compliment the code.
Language: English.
```

The system will automatically load this file and use it for the next review.

---

## 5. Development & Troubleshooting

### Local Debugging
To test the CLI locally without installing it globally:

```bash
# 1. Build
npm run build

# 2. Run CLI
# Replace <sha> with a real commit SHA from your repo
npx ts-node src/cli/index.ts run \
  --platform github \
  --repo owner/repo \
  --pr 1 \
  --sha <head_sha>
```

### Common Issues

**Q: The bot is reviewing deleted files.**
A: Ensure your `IGNORE_PATTERNS` includes `lock` or specific folders. The bot tries to skip deleted files automatically, but context is helpful.

**Q: GitLab comments are not appearing.**
A: GitLab's API requires precise `old_line` and `new_line` references. If the Diff is outdated, comments might fail. Try using `/review` to refresh.

**Q: `EPERM` or Network errors in Docker.**
A: Check if your Docker container has DNS access to the Git platform (especially for self-hosted GitLab/Gitea).
