import express from 'express';
import * as fs from 'fs';
import { config } from '../config';
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
  // Debug: Dump request to file (Handling circular references)
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

  try {
    fs.writeFileSync('req_dump.json', JSON.stringify(req, getCircularReplacer(), 2));
  } catch (e) {
    console.error('Failed to dump request', e);
  }

  const eventType =
    req.headers['x-github-event'] || req.headers['x-gitea-event'] || req.headers['x-gitlab-event'];
  console.log('Received webhook:', eventType);

  let provider;
  if (req.headers['x-github-event']) {
    provider = githubProvider;
  } else if (req.headers['x-gitea-event']) {
    provider = giteaProvider;
  } else {
    // Default or GitLab
    console.log('Unknown or unsupported platform');
    res.status(400).send('Unsupported platform');
    return;
  }

  try {
    const event = await provider.parseEvent(req);
    if (!event) {
      console.log('Event ignored (not a PR opened/sync)');
      res.status(200).send('Ignored');
      return;
    }

    // Process in background (don't block the webhook response)
    const orchestrator = new ReviewOrchestrator(provider, aiProvider);
    orchestrator.handleEvent(event).catch((err) => {
      console.error('Error handling event:', err);
    });

    res.status(200).send('Processing');
  } catch (error) {
    console.error('Webhook error:', error);
    res.status(500).send('Internal Server Error');
  }
});

const PORT = config.port;
app.listen(PORT, () => {
  console.log(`AI Code Reviewer listening on port ${PORT}`);
});
