import { Octokit } from 'octokit';
import { IPlatformProvider, ReviewEvent, FileChange, ReviewComment } from '../interfaces';
import { config } from '../config';

export class GitHubProvider implements IPlatformProvider {
  private octokit: Octokit;

  constructor() {
    this.octokit = new Octokit({ auth: config.github.token });
  }

  async parseEvent(request: any): Promise<ReviewEvent | null> {
    const payload = request.body;
    const eventName = request.headers['x-github-event'];

    if (
      eventName === 'pull_request' &&
      ['synchronize', 'reopened', 'opened'].includes(payload.action)
    ) {
      return {
        platform: 'github',
        eventType: 'pull_request',
        eventId: payload.pull_request.number.toString(),
        repoName: payload.repository.name,
        owner: payload.repository.owner.login,
        commitSha: payload.pull_request.head.sha,
        baseSha: payload.pull_request.base.sha,
        installationId: payload.installation?.id,
      };
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
}
