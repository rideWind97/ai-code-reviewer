import axios from 'axios';
import Logger from '../../utils/logger';
import config from '../../config/index.js';

const { ignoreCr } = config.review;

/**
 * GitHub API 操作服务类
 */
class GitHubAPI {
  constructor() {
    this.baseURL = 'https://api.github.com';
    this.token = config.github.token;
    this.headers = {
      Authorization: `token ${this.token}`,
      Accept: 'application/vnd.github.v3+json',
      'User-Agent': 'AI-Code-Reviewer',
    };
  }

  /**
   * 根据分支查找关联的 PR
   */
  async findPullRequestByBranch(owner, repo, branch) {
    try {
      Logger.info('查找分支关联的PR', { owner, repo, branch });

      const response = await axios.get(`${this.baseURL}/repos/${owner}/${repo}/pulls`, {
        params: { state: 'open', head: `${owner}:${branch}` },
        headers: this.headers,
        timeout: 10000,
      });

      const prs = response.data;
      if (prs && prs.length > 0) {
        const latestPR = prs.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))[0];
        Logger.info('找到关联PR', { owner, repo, branch, prNumber: latestPR.number });
        return latestPR;
      }

      Logger.warn('未找到关联的PR', { owner, repo, branch });
      return null;
    } catch (err) {
      Logger.error('查找PR失败', err, { owner, repo, branch });
      throw err;
    }
  }

  /**
   * 获取 PR 变更内容
   */
  async getPRChanges(owner, repo, prNumber) {
    try {
      const prResponse = await axios.get(
        `${this.baseURL}/repos/${owner}/${repo}/pulls/${prNumber}`,
        { headers: this.headers, timeout: 10000 },
      );

      const prInfo = prResponse.data;

      // 检查PR标题是否包含"no-cr"，如果包含则跳过代码审查
      if (prInfo.title && prInfo.title.toLowerCase().includes(ignoreCr)) {
        Logger.info(`🚫 PR标题包含"no-cr"，跳过代码审查: ${prInfo.title}`);
        return { skipReview: true, title: prInfo.title };
      }

      const filesResponse = await axios.get(
        `${this.baseURL}/repos/${owner}/${repo}/pulls/${prNumber}/files`,
        { headers: this.headers, timeout: 10000 },
      );

      const files = filesResponse.data;
      if (!files || files.length === 0) {
        return null;
      }

      const enrichedFiles = files.map((file) => ({
        ...file,
        new_path: file.filename,
        old_path: file.previous_filename || file.filename,
        diff: file.patch,
        base_sha: prInfo.base.sha,
        start_sha: prInfo.base.sha,
        head_sha: prInfo.head.sha,
      }));

      return enrichedFiles;
    } catch (err) {
      Logger.error('获取 PR 变更失败', err);
      throw err;
    }
  }

  /**
   * 发布行内评论到 PR
   */
  async postInlineCommentsToPR(owner, repo, prNumber, changes, fileReviews) {
    try {
      if (!fileReviews || fileReviews.length === 0) {
        Logger.info('没有审查内容，跳过行内评论');
        return true;
      }

      const fileComments = [];

      for (const change of changes) {
        const filePath = change.new_path || change.old_path;
        if (!filePath) continue;

        const fileReview = fileReviews.find((fr) => fr.filePath === filePath);
        if (!fileReview) continue;

        const fileComment = { filePath, change, comments: [] };

        for (const lineReview of fileReview.review) {
          if (lineReview.isGroupEnd) {
            fileComment.comments.push({
              line: lineReview.lineNumber,
              body: lineReview.review,
              position: lineReview.lineNumber,
            });
          }
        }

        if (fileComment.comments.length > 0) {
          fileComments.push(fileComment);
        }
      }

      for (const fileComment of fileComments) {
        await this.addInlineCommentsToFile(owner, repo, prNumber, fileComment);
      }

      return true;
    } catch (err) {
      Logger.error('发布行内评论失败', err);
      throw err;
    }
  }

  /**
   * 为单个文件添加行内评论
   */
  async addInlineCommentsToFile(owner, repo, prNumber, fileComment) {
    try {
      for (const comment of fileComment.comments) {
        await axios.post(
          `${this.baseURL}/repos/${owner}/${repo}/pulls/${prNumber}/comments`,
          {
            body: comment.body,
            path: fileComment.filePath,
            position: comment.position,
          },
          { headers: this.headers, timeout: 10000 },
        );

        await new Promise((resolve) => setTimeout(resolve, 100));
      }

      Logger.info(`文件 ${fileComment.filePath} 的行内评论添加完成`);
      return true;
    } catch (err) {
      Logger.error(`为文件 ${fileComment.filePath} 添加行内评论失败`, err);
      throw err;
    }
  }

  /**
   * 获取已有评论
   */
  async getExistingComments(owner, repo, prNumber) {
    try {
      const response = await axios.get(
        `${this.baseURL}/repos/${owner}/${repo}/pulls/${prNumber}/comments`,
        { headers: this.headers, timeout: 10000 },
      );

      return response.data || [];
    } catch (err) {
      Logger.error('获取已有评论失败', err);
      return [];
    }
  }

  /**
   * 添加审查评论到 PR
   */
  async postCommentToPR(owner, repo, prNumber, comment) {
    try {
      await axios.post(
        `${this.baseURL}/repos/${owner}/${repo}/issues/${prNumber}/comments`,
        { body: comment },
        { headers: this.headers, timeout: 10000 },
      );

      Logger.info('评论发布成功');
      return true;
    } catch (err) {
      Logger.error('发布评论失败', err);
      throw err;
    }
  }
}

export default GitHubAPI;
