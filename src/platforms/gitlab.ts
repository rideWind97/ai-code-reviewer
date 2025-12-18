import { IPlatformProvider, ReviewEvent, FileChange, ReviewComment } from '../interfaces';
import { config } from '../config';
import { Gitlab } from '@gitbeaker/rest';

export class GitLabProvider implements IPlatformProvider {
  private api: InstanceType<typeof Gitlab>;

  constructor() {
    this.api = new Gitlab({
      host: config.gitlab.host,
      token: config.gitlab.token,
    });
  }

  async parseEvent(request: any): Promise<ReviewEvent | null> {
    const payload = request.body;
    const objectKind = payload.object_kind;

    if (
      objectKind === 'merge_request' &&
      (payload.object_attributes.action === 'open' || payload.object_attributes.action === 'update')
    ) {
      const attrs = payload.object_attributes;
      return {
        platform: 'gitlab',
        eventType: 'merge_request',
        eventId: attrs.iid.toString(),
        projectId: payload.project.id,
        repoName: payload.project.path_with_namespace,
        owner: payload.project.namespace,
        commitSha: attrs.last_commit.id,
        baseSha: attrs.target.id || payload.diff_refs?.base_sha, // Attempt to get Base SHA
      };
    }
    return null;
  }

  async getChanges(event: ReviewEvent): Promise<FileChange[]> {
    if (!event.projectId) return [];

    try {
      // Fetch changes from GitLab
      // Using mergeRequests.changes to get the diffs
      const changes = await this.api.MergeRequests.showChanges(
        event.projectId,
        parseInt(event.eventId),
      );

      return (changes.changes as any[]).map((change: any) => ({
        filename: change.new_path || change.newPath,
        status:
          change.new_file || change.newFile
            ? 'added'
            : change.deleted_file || change.deletedFile
              ? 'deleted'
              : 'modified',
        rawDiff: change.diff,
        blobUrl: '', // Not easily available in changes response, can be constructed if needed
      }));
    } catch (e) {
      console.error('GitLab getChanges error:', e);
      return [];
    }
  }

  async postReview(event: ReviewEvent, comments: ReviewComment[]): Promise<void> {
    if (!event.projectId || comments.length === 0) return;

    // GitLab API for inline comments (Discussions)
    // We need to create a discussion for each comment (or group them).
    // The difficult part is the 'position' object.

    for (const comment of comments) {
      try {
        await this.api.MergeRequestDiscussions.create(
          event.projectId,
          parseInt(event.eventId),
          comment.body,
          {
            position: {
              baseSha: event.baseSha,
              startSha: event.baseSha, // Usually same as base for MRs
              headSha: event.commitSha,
              positionType: 'text',
              newPath: comment.path,
              newLine: comment.line.toString(),
            },
          },
        );
      } catch (e) {
        console.error(`Failed to post GitLab comment on ${comment.path}:${comment.line}`, e);
      }
    }
  }
}
