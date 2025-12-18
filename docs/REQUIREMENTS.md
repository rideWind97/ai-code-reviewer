# Product Requirements Document (PRD)

[🇨🇳 中文](./REQUIREMENTS.zh-CN.md) | **🇺🇸 English**

## 1. Project Overview

A Node.js-based tool that utilizes Large Language Models (LLM) to automatically review code changes (Pull Requests/Merge Requests) across multiple Git hosting platforms. It aims to catch potential bugs, style issues, and security vulnerabilities before human review.

## 2. Key Features

### 2.1 Multi-Platform Support

The system must support the following platforms via a unified abstraction layer:

- **GitHub** (Cloud & Enterprise)
- **GitLab** (Cloud & Self-hosted)
- **Gitea** (Self-hosted)

### 2.2 Integration Modes

- **Webhook Server Mode**: Listens for `pull_request` or `merge_request` events and triggers reviews automatically.
- **CI/CD CLI Mode**: Can be run as a step in a pipeline (e.g., GitHub Actions, GitLab CI) to review the current commit/PR.

### 2.3 Code Analysis Capabilities

- **Diff Analysis**: Fetches the diff of the PR, parsing added/modified files.
- **Context Awareness**: Understands file types and provides language-specific advice.
- **Filtering**: Ignores generated files (e.g., `package-lock.json`, `dist/`), binary files, or specific patterns defined in configuration.

### 2.4 Feedback Mechanism

- **Inline Comments**: Specific comments on the exact line where the issue is detected.
- **General Summary**: A top-level comment summarizing the overall quality and security score of the PR.

## 3. Configuration & Customization

Users should be able to configure via a config file (e.g., `.aicrrc` or env vars):

- **Target Branch**: Which branches to review (e.g., only `main` or `develop`).
- **Review Language**: Output language (Chinese/English).
- **Prompt Customization**: Ability to tweak the system prompt for the AI.
- **Severity Level**: Filter out "nitpick" comments if desired.

## 4. Security Requirements

- **Secrets Management**: API keys (OpenAI, GitHub Token) must never be logged.
- **Code Privacy**: Code snippets sent to AI must be minimized to the diff context.

## 5. User Stories

- **Developer**: I want my PR to get immediate feedback on common mistakes so I can fix them before assigning a human reviewer.
- **Team Lead**: I want to enforce coding standards automatically across the repo.
- **DevOps Engineer**: I want to easily deploy this as a Docker container or a GitHub Action.
