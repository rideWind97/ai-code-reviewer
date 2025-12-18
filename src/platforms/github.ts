import { Octokit } from 'octokit';
import { IPlatformProvider, ReviewEvent, FileChange, ReviewComment } from '../interfaces';
import { config } from '../config';
import { logger } from '../utils/logger';

export class GitHubProvider implements IPlatformProvider {
  private octokit: Octokit;

  constructor() {
    this.octokit = new Octokit({ auth: config.github.token });
  }

  async parseEvent(request: any): Promise<ReviewEvent | null> {
    const payload = request.body;
    const eventName = request.headers['x-github-event'];

    if (eventName === 'pull_request' && ['synchronize', 'opened'].includes(payload.action)) {
      return {
        platform: 'github',
        eventType: 'pull_request',
        eventId: payload.pull_request.number.toString(),
        repoName: payload.repository.name,
        owner: payload.repository.owner.login,
        commitSha: payload.pull_request.head.sha,
        baseSha: payload.pull_request.base.sha,
        installationId: payload.installation?.id,
        isDraft: payload.pull_request.draft,
      };
    }

    // ChatOps: Handle /review command in comments
    if (eventName === 'issue_comment' && payload.action === 'created') {
      const body = payload.comment.body || '';
      if (body.trim().startsWith('/review') && payload.issue.pull_request) {
        logger.info(
          { issue: payload.issue.number, repo: payload.repository.name },
          'Received /review command',
        );

        try {
          const { data: pr } = await this.octokit.rest.pulls.get({
            owner: payload.repository.owner.login,
            repo: payload.repository.name,
            pull_number: payload.issue.number,
          });

          return {
            platform: 'github',
            eventType: 'pull_request',
            eventId: pr.number.toString(),
            repoName: payload.repository.name,
            owner: payload.repository.owner.login,
            commitSha: pr.head.sha,
            baseSha: pr.base.sha,
            installationId: payload.installation?.id,
            isDraft: pr.draft,
          };
        } catch (error) {
          logger.error({ err: error }, 'Failed to fetch PR details for ChatOps');
        }
      }
    }

    return null;
  }

  async getChanges(event: ReviewEvent): Promise<FileChange[]> {
    const { data: files } = await this.octokit.rest.pulls.listFiles({
      owner: event.owner,
      repo: event.repoName,
      pull_number: parseInt(event.eventId),
    });

    return files.map((file: any) => ({
      filename: file.filename,
      status: file.status as any,
      rawDiff: file.patch || '',
      blobUrl: file.blob_url,
    }));
  }

  async postReview(event: ReviewEvent, comments: ReviewComment[]): Promise<void> {
    if (comments.length === 0) return;

    const ghComments = comments.map((c) => ({
      path: c.path,
      line: c.line,
      body: c.body,
    }));

    await this.octokit.rest.pulls.createReview({
      owner: event.owner,
      repo: event.repoName,
      pull_number: parseInt(event.eventId),
      commit_id: event.commitSha,
      event: 'COMMENT',
      comments: ghComments,
    });
  }

  async getReviewComments(event: ReviewEvent): Promise<ReviewComment[]> {
    try {
      const { data } = await this.octokit.rest.pulls.listReviewComments({
        owner: event.owner,
        repo: event.repoName,
        pull_number: parseInt(event.eventId),
      });

      return data.map((c: any) => ({
        path: c.path,
        line: c.line || c.original_line || 0,
        body: c.body,
        side: c.side as 'LEFT' | 'RIGHT',
      }));
    } catch (error) {
      logger.error({ err: error }, 'Error fetching GitHub comments');
      return [];
    }
  }
}
