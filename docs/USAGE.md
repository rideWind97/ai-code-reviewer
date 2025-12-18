# Usage Guide

[🇨🇳 中文](./USAGE.zh-CN.md) | **🇺🇸 English**

This document details how to configure, deploy, and use the AI Code Reviewer.

## 1. Environment Configuration (.env)

Create a `.env` file in the project root:

```ini
# --- Service Configuration ---
PORT=4000
NODE_ENV=production

# --- AI Service Configuration ---
AI_PROVIDER=openai
# OpenAI API Key (Required)
AI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxx
# Model Name (Recommended: gpt-4o, gpt-3.5-turbo)
AI_MODEL=gpt-4o
# Optional: Custom API URL (For proxies or compatible models)
AI_BASE_URL=https://api.openai.com/v1

# --- Git Platform Configuration (Fill as needed) ---

# GitHub
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxx    # Personal Access Token (PAT)
GITHUB_WEBHOOK_SECRET=your_secret    # Webhook Secret

# GitLab
GITLAB_HOST=https://gitlab.com       # Change if self-hosted
GITLAB_TOKEN=glpat-xxxxxxxxxxxx      # Access Token

# Gitea
GITEA_HOST=https://gitea.example.com
GITEA_TOKEN=xxxxxxxxxxxxxxxxxxxx

# --- Review Rules ---
# Target Branches (Comma separated)
TARGET_BRANCHES=main,master,develop
# Ignored file patterns
IGNORE_PATTERNS=package-lock.json,dist,node_modules,yarn.lock
# Review Language
REVIEW_LANGUAGE=en
# Skip Draft PR
SKIP_DRAFT_PR=true
# Concurrency Limit
CONCURRENCY_LIMIT=5
```

## 2. Deployment Modes

### Mode A: Webhook Server (Recommended)

Run as a standalone service, listening for Webhook events in real-time.

1.  **Start Service**:

    ```bash
    docker-compose up -d
    ```

    Service runs on port `4000` by default.

2.  **Configure Webhook**:
    - **GitHub**: Repository Settings -> Webhooks -> Add webhook
      - Payload URL: `http://your-server-ip:4000/webhook`
      - Content type: `application/json`
      - Secret: (Match .env `GITHUB_WEBHOOK_SECRET`)
      - Events: Select `Pull requests`
    - **GitLab**: Settings -> Webhooks
      - URL: `http://your-server-ip:4000/webhook`
      - Trigger: `Merge request events`
    - **Gitea**: Settings -> Webhooks
      - Target URL: `http://your-server-ip:4000/webhook`
      - Events: `Pull Request`

### Mode B: CLI / CI Integration

Run as a step in your CI/CD pipeline.

## 3. ChatOps (Interactive Commands)

In addition to automatic PR event listening, you can manually trigger reviews by typing commands in the PR/MR comment section (Issue Comments):

- `/review`: Re-review the current PR.

Note: Ensure **Issue comments** (GitHub/Gitea) or **Note events** (GitLab) are enabled in your Webhook settings.

#### 1. GitHub Actions

Add to your `.github/workflows/ai-review.yml`:

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
        run: npx ts-node src/cli/index.ts run --platform github
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          AI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          # GITHUB_EVENT_PATH is injected by Actions automatically
```

#### 2. GitLab CI

In `.gitlab-ci.yml`:

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

#### 3. Manual Execution (CLI)

You can trigger a review locally:

```bash
# Review GitHub PR #123
npx ts-node src/cli/index.ts run \
  --platform github \
  --repo owner/repo \
  --pr 123 \
  --sha <head_commit_sha>
```

## 4. Development

### Debug CLI Locally

```bash
# Build
npm run build

# Link Command (Optional)
npm link

# Run
aicr run --help
```

### FAQ

1.  **GitLab Comment Failed**:
    - Ensure `GITLAB_HOST` and `GITLAB_TOKEN` are correct.
    - GitLab inline comments require exact `base_sha`, `head_sha`, and `start_sha`. Provide `--base-sha` in CLI mode.

2.  **Slow AI Response**:
    - Check Diff size. Use `IGNORE_PATTERNS` to filter large files.
    - OpenAI API latency. Try compatible local models or proxies.
