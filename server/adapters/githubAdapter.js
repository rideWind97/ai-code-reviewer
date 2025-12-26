import axios from 'axios';
import GitPlatformAdapter from '../core/interfaces/gitPlatform';
import CodeChange from '../core/models/codeChange';
import ReviewComment from '../core/models/reviewComment';
import Logger from '../utils/logger';

export default class GitHubAdapter extends GitPlatformAdapter {
  constructor(config) {
    super(config);
    this.token = config.github.token;
    this.timeout = config.github.timeout || 10000;
  }

  get headers() {
    return {
      Authorization: `token ${this.token}`,
      Accept: 'application/vnd.github.v3+json',
      'User-Agent': 'AI-Code-Reviewer',
    };
  }

  parseProjectId(projectId) {
    // projectId should be "owner/repo"
    const [owner, repo] = String(projectId).split('/');
    if (!owner || !repo) {
      throw new Error(`Invalid GitHub projectId: ${projectId}. Expected format: "owner/repo"`);
    }
    return { owner, repo };
  }

  /**
   * @override
   * @param {import('../core/models/reviewRequest')} reviewRequest
   */
  async getChanges(reviewRequest) {
    const { projectId, mergeRequestId } = reviewRequest;
    const { owner, repo } = this.parseProjectId(projectId);
    const prNumber = mergeRequestId;

    try {
      // 获取 PR 详情以获取 SHA 信息
      const prInfoUrl = `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}`;
      const prResponse = await axios.get(prInfoUrl, {
        headers: this.headers,
        timeout: this.timeout,
      });
      const prInfo = prResponse.data;

      // 获取 PR 文件列表
      const filesUrl = `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}/files`;
      const response = await axios.get(filesUrl, { headers: this.headers, timeout: this.timeout });

      return (response.data || []).map(
        (file) =>
          new CodeChange({
            filePath: file.filename,
            oldPath: file.previous_filename,
            content: file.patch || '', // GitHub patch property contains the diff
            isNewFile: file.status === 'added',
            isDeleted: file.status === 'removed',
            baseSha: prInfo.base.sha,
            headSha: prInfo.head.sha,
          }),
      );
    } catch (err) {
      Logger.error(`GitHub getChanges failed for ${projectId} PR ${prNumber}`, err.message);
      return [];
    }
  }

  /**
   * @override
   * @param {import('../core/models/reviewRequest')} reviewRequest
   */
  async getExistingComments(reviewRequest) {
    const { projectId, mergeRequestId } = reviewRequest;
    const { owner, repo } = this.parseProjectId(projectId);
    const prNumber = mergeRequestId;

    try {
      // GitHub PR Review Comments (inline comments)
      const url = `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}/comments`;
      const response = await axios.get(url, { headers: this.headers, timeout: this.timeout });

      return (response.data || []).map(
        (comment) =>
          new ReviewComment({
            filePath: comment.path,
            lineNumber: comment.line || comment.original_line, // GitHub uses 'line' for current view, 'original_line' for older
            content: comment.body,
            side: comment.side === 'LEFT' ? 'LEFT' : 'RIGHT',
          }),
      );
    } catch (err) {
      Logger.error(
        `GitHub getExistingComments failed for ${projectId} PR ${prNumber}`,
        err.message,
      );
      return [];
    }
  }

  /**
   * @override
   * @param {import('../core/models/reviewRequest')} reviewRequest
   * @param {import('../core/models/reviewComment')[]} comments
   */
  async publishComments(reviewRequest, comments) {
    const { projectId, mergeRequestId, commitSha } = reviewRequest;
    const { owner, repo } = this.parseProjectId(projectId);
    const prNumber = mergeRequestId;

    let successCount = 0;
    let failCount = 0;

    // GitHub 建议使用 Review API 批量提交评论，或者是逐个提交
    // 这里先实现逐个提交，后续可优化为 createReview 批量提交

    for (const comment of comments) {
      try {
        const url = `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}/comments`;

        // GitHub API requires 'commit_id', 'path', 'body', 'line' (or 'position' deprecated)
        // 注意：commit_id 必须是 PR head 的 SHA

        const payload = {
          body: comment.content,
          path: comment.filePath,
          line: comment.lineNumber,
          side: 'RIGHT', // 默认针对新代码
          commit_id: commitSha || (await this.getLatestCommitSha(owner, repo, prNumber)),
        };

        await axios.post(url, payload, { headers: this.headers, timeout: this.timeout });
        successCount++;
        await new Promise((resolve) => setTimeout(resolve, 100)); // Rate limiting
      } catch (err) {
        Logger.error(
          `GitHub publish comment failed ${comment.filePath}:${comment.lineNumber}`,
          err.message,
        );
        failCount++;
      }
    }

    // 如果有评论，还可以发一个总的 Summary (Issue Comment)
    if (comments.length > 0 && this.config.review.addSummaryComment) {
      await this.postSummary(owner, repo, prNumber, comments.length);
    }

    return { success: successCount, failed: failCount };
  }

  async getLatestCommitSha(owner, repo, prNumber) {
    try {
      const url = `https://api.github.com/repos/${owner}/${repo}/pulls/${prNumber}`;
      const response = await axios.get(url, { headers: this.headers, timeout: this.timeout });
      return response.data.head.sha;
    } catch (err) {
      return null;
    }
  }

  async postSummary(owner, repo, prNumber, count) {
    try {
      const url = `https://api.github.com/repos/${owner}/${repo}/issues/${prNumber}/comments`;
      const body = `## 🤖 AI Code Review\n\nCompleted review. Found ${count} potential issues.`;
      await axios.post(url, { body }, { headers: this.headers, timeout: this.timeout });
    } catch (err) {
      Logger.warn('Failed to post summary to GitHub', err.message);
    }
  }

  /**
   * @override
   * 获取项目文件内容
   */
  async getProjectFile(projectId, filePath, ref = 'main') {
    try {
      const { owner, repo } = this.parseProjectId(projectId);
      const url = `https://api.github.com/repos/${owner}/${repo}/contents/${filePath}?ref=${ref}`;
      const response = await axios.get(url, {
        headers: { ...this.headers, Accept: 'application/vnd.github.v3.raw' },
        timeout: this.timeout,
      });
      return response.data; // raw content
    } catch (err) {
      if (err.response && err.response.status === 404) {
        return null;
      }
      Logger.error(`GitHub getProjectFile failed for ${projectId} file ${filePath}`, err.message);
      return null;
    }
  }
}
