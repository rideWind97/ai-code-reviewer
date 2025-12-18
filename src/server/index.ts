import express from 'express';
import * as fs from 'fs';
import { config } from '../config';
import { logger } from '../utils/logger';
import { GitHubProvider } from '../platforms/github';
import { GiteaProvider } from '../platforms/gitea';
import { OpenAIProvider } from '../ai/openai';
import { ReviewOrchestrator } from '../core/orchestrator';

const app = express();
app.use(express.json());

// Initialize Providers
// In a real app, you might want to lazily init these or use dependency injection
const githubProvider = new GitHubProvider();
const giteaProvider = new GiteaProvider();
// const gitlabProvider = new GitLabProvider();

const aiProvider = new OpenAIProvider();

app.get('/ping', (req, res) => {
  res.status(200).send('pong');
});

app.post('/webhook', async (req, res) => {
  // Debug: Dump request to file (Only in development)
  if (process.env.NODE_ENV === 'development') {
    const getCircularReplacer = () => {
      const seen = new WeakSet();
      return (key: any, value: any) => {
        if (typeof value === 'object' && value !== null) {
          if (seen.has(value)) {
            return '[Circular]';
          }
          seen.add(value);
        }
        return value;
      };
    };

    fs.promises
      .writeFile('req_dump.json', JSON.stringify(req, getCircularReplacer(), 2))
      .catch((e) => console.error('Failed to dump request', e));
  }

  const eventType =
    req.headers['x-github-event'] || req.headers['x-gitea-event'] || req.headers['x-gitlab-event'];
  logger.info({ eventType }, 'Received webhook');

  let provider;
  if (req.headers['x-github-event']) {
    provider = githubProvider;
  } else if (req.headers['x-gitea-event']) {
    provider = giteaProvider;
  } else {
    // Default or GitLab
    logger.warn('Unknown or unsupported platform');
    res.status(400).send('Unsupported platform');
    return;
  }

  try {
    const event = await provider.parseEvent(req);
    if (!event) {
      logger.info('Event ignored (not a PR opened/sync)');
      res.status(200).send('Ignored');
      return;
    }

    // Process in background (don't block the webhook response)
    const orchestrator = new ReviewOrchestrator(provider, aiProvider);
    orchestrator.handleEvent(event).catch((err) => {
      logger.error({ err }, 'Error handling event');
    });

    res.status(200).send('Processing');
  } catch (error) {
    logger.error({ error }, 'Webhook error');
    res.status(500).send('Internal Server Error');
  }
});

const PORT = config.port;
app.listen(PORT, () => {
  logger.info(`AI Code Reviewer listening on port ${PORT}`);
});
