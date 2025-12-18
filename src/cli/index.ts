#!/usr/bin/env node
import { Command } from 'commander';
import { GitHubProvider } from '../platforms/github';
import { GitLabProvider } from '../platforms/gitlab';
import { GiteaProvider } from '../platforms/gitea';
import { OpenAIProvider } from '../ai/openai';
import { ReviewOrchestrator } from '../core/orchestrator';
import { ReviewEvent } from '../interfaces';
import * as fs from 'fs';
import { logger } from '../utils/logger';

const program = new Command();

program
  .name('ai-code-reviewer')
  .description('AI-powered code review tool for GitHub, GitLab, and Gitea')
  .version('1.0.0');

program
  .command('run')
  .description('Run code review manually or in CI')
  .option('-p, --platform <type>', 'Platform type (github, gitlab, gitea)', 'github')
  .option('--event-path <path>', 'Path to webhook event JSON (useful for GitHub Actions)')
  .option('--pr <number>', 'Pull/Merge Request Number')
  .option('--repo <string>', 'Repository name (owner/repo)')
  .option('--sha <string>', 'Commit SHA to review')
  .option('--base-sha <string>', 'Base SHA for diff comparison')
  .action(async (options) => {
    try {
      logger.info('Starting AI Code Review...');

      let event: ReviewEvent | null = null;
      let provider;
      const aiProvider = new OpenAIProvider();

      // 1. Initialize Provider
      switch (options.platform) {
        case 'github':
          provider = new GitHubProvider();
          break;
        case 'gitlab':
          provider = new GitLabProvider();
          break;
        case 'gitea':
          provider = new GiteaProvider();
          break;
        default:
          throw new Error(`Unsupported platform: ${options.platform}`);
      }

      // 2. Construct Event
      // Scenario A: GitHub Actions (using event.json)
      if (options.platform === 'github' && process.env.GITHUB_EVENT_PATH) {
        const eventData = JSON.parse(fs.readFileSync(process.env.GITHUB_EVENT_PATH, 'utf8'));
        if (eventData.pull_request) {
          event = {
            platform: 'github',
            eventType: 'pull_request',
            eventId: eventData.pull_request.number.toString(),
            repoName: eventData.repository.full_name,
            owner: eventData.repository.owner.login,
            commitSha: eventData.pull_request.head.sha,
            baseSha: eventData.pull_request.base.sha,
            installationId: undefined, // Actions usually use GITHUB_TOKEN directly, not App ID
          };
        }
      }
      // Scenario B: Manual / Flags
      else if (options.pr && options.repo) {
        if (!options.sha)
          console.warn('Warning: --sha not provided, might fail for some platforms');
        if (!options.baseSha)
          console.warn('Warning: --base-sha not provided, inline comments might fail on GitLab');

        event = {
          platform: options.platform,
          eventType: options.platform === 'gitlab' ? 'merge_request' : 'pull_request',
          eventId: options.pr,
          repoName: options.repo,
          owner: options.repo.split('/')[0],
          commitSha: options.sha || '', // Logic might need to fetch if missing
          baseSha: options.baseSha || '',
          projectId: options.repo, // GitLab often accepts path as ID
        };
      }

      if (!event) {
        console.error(
          'Error: Could not determine review event context. Please provide --pr and --repo, or run inside a supported CI environment.',
        );
        process.exit(1);
      }

      // 3. Run Orchestrator
      const orchestrator = new ReviewOrchestrator(provider, aiProvider);
      await orchestrator.handleEvent(event);

      console.log('Review completed successfully.');
    } catch (error) {
      console.error('Fatal Error:', error);
      process.exit(1);
    }
  });

program.parse(process.argv);
