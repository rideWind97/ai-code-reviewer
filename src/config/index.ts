import * as dotenv from 'dotenv';
dotenv.config();

export const config = {
  port: process.env.PORT || 3000,
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
  },
  review: {
    targetBranches: (process.env.TARGET_BRANCHES || 'main,master,develop').split(','),
    ignorePatterns: (process.env.IGNORE_PATTERNS || 'lock,dist,node_modules').split(','),
  },
};
