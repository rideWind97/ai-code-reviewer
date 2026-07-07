# AICR

AICR is a Go service skeleton for automated GitHub pull request code review. The MVP focuses on a review-only loop: receive a GitHub webhook, build review units from the diff, call a model, aggregate findings, and publish comments/status back through the provider adapter.

## Current MVP Scope

- GitHub pull request webhook verification and idempotent queueing.
- Provider-neutral patch/comment/status interfaces with a GitHub adapter.
- Review run state machine and an in-process runner.
- Review unit chunking, context pack building, model gateway, finding aggregation, publishing helpers, and secret scanning.
- Local PostgreSQL schema under `migrations/`.

GitLab support, persistent queue workers, GitHub App installation token exchange, repository cache, large PR summaries, and project-level review skills are planned follow-ups.

## Configuration

Environment variables:

- `AICR_ADDR`: HTTP listen address, default `:8080`.
- `AICR_DATABASE_URL`: PostgreSQL connection string.
- `AICR_GITHUB_WEBHOOK_SECRET`: GitHub webhook secret.
- `AICR_GITHUB_TOKEN`: GitHub API token for the current adapter skeleton.
- `AICR_MODEL_BASE_URL`: OpenAI-compatible API base URL.
- `AICR_MODEL_NAME`: model name.
- `AICR_MODEL_API_KEY`: model API key.

## Development

Start Postgres:

```bash
docker compose up -d postgres
```

Run tests:

```bash
make test
```

Run the HTTP server:

```bash
make run-server
```

Run the worker placeholder:

```bash
make run-worker
```

## Architecture Notes

The implementation follows `docs/ai-coding-agent-architecture.md` as a modular monolith. Packages under `internal/review` keep the review pipeline decoupled from provider-specific code, while `internal/provider/github` owns GitHub API details.
