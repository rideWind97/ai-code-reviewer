import { IPlatformProvider, ReviewEvent, FileChange, ReviewComment } from '../interfaces';
import { config } from '../config';
import { logger } from '../utils/logger';
import axios from 'axios';

export class GiteaProvider implements IPlatformProvider {
  private baseUrl: string;
  private token: string;

  constructor() {
    this.baseUrl = config.gitea.host || '';
    this.token = config.gitea.token || '';
  }

  async parseEvent(request: any): Promise<ReviewEvent | null> {
    const payload = request.body;
    const eventName = request.headers['x-gitea-event'];

    if (
      eventName === 'pull_request' &&
      (payload.action === 'opened' || payload.action === 'synchronized')
    ) {
      return {
        platform: 'gitea',
        eventType: 'pull_request',
        eventId: payload.pull_request.number.toString(),
        repoName: payload.repository.name,
        owner: payload.repository.owner.login,
        commitSha: payload.pull_request.head.sha,
        baseSha: payload.pull_request.base.sha,
        isDraft: payload.pull_request.work_in_progress, // Gitea WIP status
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
          // Fetch PR details to get head SHA
          const url = `${this.baseUrl}/api/v1/repos/${payload.repository.owner.login}/${payload.repository.name}/pulls/${payload.issue.number}`;
          const { data: pr } = await axios.get(url, {
            headers: { Authorization: `token ${this.token}` },
          });

          return {
            platform: 'gitea',
            eventType: 'pull_request',
            eventId: pr.number.toString(),
            repoName: payload.repository.name,
            owner: payload.repository.owner.login,
            commitSha: pr.head.sha,
            baseSha: pr.base.sha,
            isDraft: pr.work_in_progress,
          };
        } catch (error) {
          logger.error({ err: error }, 'Failed to fetch Gitea PR details for ChatOps');
        }
      }
    }

    return null;
  }

  async getChanges(event: ReviewEvent): Promise<FileChange[]> {
    const url = `${this.baseUrl}/api/v1/repos/${event.owner}/${event.repoName}/pulls/${event.eventId}/files`;

    try {
      const { data } = await axios.get(url, {
        headers: { Authorization: `token ${this.token}` },
      });

      return data.map((file: any) => ({
        filename: file.filename,
        status: file.status,
        rawDiff: file.patch || '',
        blobUrl: file.raw_url,
      }));
    } catch (error) {
      logger.error({ err: error }, 'Error fetching Gitea changes');
      return [];
    }
  }

  async postReview(event: ReviewEvent, comments: ReviewComment[]): Promise<void> {
    if (comments.length === 0) return;

    const url = `${this.baseUrl}/api/v1/repos/${event.owner}/${event.repoName}/pulls/${event.eventId}/reviews`;

    const giteaComments = comments.map((c) => ({
      path: c.path,
      new_line: c.line,
      body: c.body,
    }));

    try {
      await axios.post(
        url,
        {
          commit_id: event.commitSha,
          body: 'AI Code Review Summary',
          event: 'COMMENT',
          comments: giteaComments,
        },
        {
          headers: { Authorization: `token ${this.token}` },
        },
      );
    } catch (error) {
      console.error('Error posting Gitea review:', error);
    }
  }

  async getReviewComments(event: ReviewEvent): Promise<ReviewComment[]> {
    return [];
  }
}
