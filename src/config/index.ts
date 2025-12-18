import * as dotenv from 'dotenv';
import * as fs from 'fs';
import * as path from 'path';

dotenv.config();

function loadSystemPrompt(): string | undefined {
  const promptPath = process.env.AI_PROMPT_PATH || 'prompt.md';
  const resolvedPath = path.resolve(process.cwd(), promptPath);

  if (fs.existsSync(resolvedPath)) {
    try {
      return fs.readFileSync(resolvedPath, 'utf-8');
    } catch (error) {
      console.warn(`Warning: Failed to read prompt file at ${resolvedPath}`, error);
    }
  }

  return process.env.AI_SYSTEM_PROMPT;
}

export const config = {
  port: process.env.PORT || 4000,
  github: {
    token: process.env.GITHUB_TOKEN,
    webhookSecret: process.env.GITHUB_WEBHOOK_SECRET,
  },
  gitlab: {
    token: process.env.GITLAB_TOKEN,
    host: process.env.GITLAB_HOST || 'https://gitlab.com',
  },
  gitea: {
    token: process.env.GITEA_TOKEN,
    host: process.env.GITEA_HOST,
  },
  ai: {
    provider: process.env.AI_PROVIDER || 'openai', // openai, anthropic, etc.
    apiKey: process.env.AI_API_KEY,
    model: process.env.AI_MODEL || 'gpt-4o',
    baseUrl: process.env.AI_BASE_URL, // For proxying or local models
    systemPrompt: loadSystemPrompt(),
  },
  review: {
    targetBranches: (process.env.TARGET_BRANCHES || 'main,master,develop').split(','),
    ignorePatterns: (process.env.IGNORE_PATTERNS || 'lock,dist,node_modules').split(','),
  },
};
