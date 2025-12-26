import axios from 'axios';
import GitPlatformAdapter from '../core/interfaces/gitPlatform';
import CodeChange from '../core/models/codeChange';
import ReviewComment from '../core/models/reviewComment';
import Logger from '../utils/logger';

export default class GitLabAdapter extends GitPlatformAdapter {
  constructor(config) {
    super(config);
    this.baseURL = config.gitlab.url;
    this.token = config.gitlab.token;
    this.timeout = config.gitlab.timeout || 10000;
  }

  get headers() {
    return {
      'PRIVATE-TOKEN': this.token,
    };
  }

  /**
   * @override
   * @param {import('../core/models/reviewRequest')} reviewRequest
   */
  async getChanges(reviewRequest) {
    const { projectId, mergeRequestId, action, commitSha } = reviewRequest;

    // MR 更新且有特定 Commit 时，优先只获取该 Commit 的变更
    if (action === 'update' && commitSha) {
      try {
        return await this.getCommitChanges(projectId, commitSha);
      } catch (err) {
        Logger.warn(`获取 Commit ${commitSha} 变更失败，降级为获取完整 MR 变更`, err.message);
      }
    }

    // 默认获取 MR 变更
    return await this.getMRChanges(projectId, mergeRequestId);
  }

  async getMRChanges(projectId, mrIid) {
    const url = `${this.baseURL}/api/v4/projects/${projectId}/merge_requests/${mrIid}/changes`;
    const response = await axios.get(url, { headers: this.headers, timeout: this.timeout });
    const changes = response.data.changes || [];
    const mrInfo = response.data;

    return changes.map(
      (change) =>
        new CodeChange({
          filePath: change.new_path,
          oldPath: change.old_path,
          content: change.diff,
          isNewFile: change.new_file,
          isDeleted: change.deleted_file,
          baseSha: mrInfo.diff_refs?.base_sha,
          headSha: mrInfo.diff_refs?.head_sha,
        }),
    );
  }

  async getCommitChanges(projectId, sha) {
    const url = `${this.baseURL}/api/v4/projects/${projectId}/repository/commits/${sha}/diff`;
    const response = await axios.get(url, { headers: this.headers, timeout: this.timeout });

    return response.data.map(
      (diff) =>
        new CodeChange({
          filePath: diff.new_path,
          oldPath: diff.old_path,
          content: diff.diff,
          isNewFile: diff.new_file,
          isDeleted: diff.deleted_file,
          headSha: sha, // Commit 模式下，head 就是当前 commit
        }),
    );
  }

  /**
   * @override
   * @param {import('../core/models/reviewRequest')} reviewRequest
   */
  async getExistingComments(reviewRequest) {
    const { projectId, mergeRequestId } = reviewRequest;
    let allNotes = [];
    let page = 1;
    const perPage = 100;

    // 简化版分页获取，实际生产建议复用完整的分页逻辑
    while (true) {
      const url = `${this.baseURL}/api/v4/projects/${projectId}/merge_requests/${mergeRequestId}/notes`;
      const response = await axios.get(url, {
        headers: this.headers,
        timeout: this.timeout,
        params: { page, per_page: perPage },
      });

      const notes = response.data;
      allNotes = allNotes.concat(notes);

      if (notes.length < perPage) break;
      page++;
    }

    return allNotes
      .filter((note) => note.position && note.position.new_line) // 只保留行内评论
      .map(
        (note) =>
          new ReviewComment({
            filePath: note.position.new_path,
            lineNumber: note.position.new_line,
            content: note.body,
            side: 'RIGHT', // GitLab API 默认也是新代码侧
          }),
      );
  }

  /**
   * @override
   * @param {import('../core/models/reviewRequest')} reviewRequest
   * @param {import('../core/models/reviewComment')[]} comments
   */
  async publishComments(reviewRequest, comments) {
    const { projectId, mergeRequestId, commitSha } = reviewRequest;
    const url = `${this.baseURL}/api/v4/projects/${projectId}/merge_requests/${mergeRequestId}/discussions`;

    let successCount = 0;
    let failCount = 0;

    for (const comment of comments) {
      try {
        const payload = {
          body: comment.content,
          position: {
            base_sha: null, // MR Context 下通常不需要手动填 base_sha，除非是特定版本对比
            start_sha: null,
            head_sha: commitSha || null, // 如果有特定 commit，最好带上
            position_type: 'text',
            new_path: comment.filePath,
            new_line: comment.lineNumber,
          },
        };

        await axios.post(url, payload, { headers: this.headers, timeout: this.timeout });
        successCount++;
        // 简单的限流
        await new Promise((resolve) => setTimeout(resolve, 100));
      } catch (err) {
        Logger.error(`发布评论失败 ${comment.filePath}:${comment.lineNumber}`, err.message);
        failCount++;
      }
    }

    return { success: successCount, failed: failCount };
  }

  /**
   * @override
   * 获取项目文件内容
   */
  async getProjectFile(projectId, filePath, ref = 'main') {
    try {
      const url = `${this.baseURL}/api/v4/projects/${projectId}/repository/files/${encodeURIComponent(filePath)}/raw?ref=${ref}`;
      const response = await axios.get(url, {
        headers: this.headers,
        timeout: this.timeout,
      });
      return response.data;
    } catch (err) {
      if (err.response && err.response.status === 404) {
        Logger.info(`项目文件不存在: ${filePath}`);
        return null;
      }
      Logger.error(`获取项目文件失败: ${filePath}`, err.message);
      return null;
    }
  }
}
