import Logger from '../utils/logger';
import config from '../config/index.js';
import GitHubAdapter from '../adapters/githubAdapter';
import CodeReviewService from '../services/codeReviewService';
import AIApiService from '../services/aiApiService';
import PromptService from '../services/promptService';
import ReviewRequest from '../core/models/reviewRequest';

/**
 * GitHub Webhook 处理器 - 新架构适配版
 */
export default class GitHubEventHandler {
  constructor() {
    this.config = config;
    this.gitAdapter = new GitHubAdapter(config);
    this.promptService = new PromptService(this.gitAdapter);
    this.aiService = new AIApiService(config.ai, this.promptService);
    this.reviewService = new CodeReviewService(this.gitAdapter, this.aiService, config);
  }

  /**
   * 处理 Push 事件
   */
  async handlePushEvent(event) {
    const { repository, ref, after } = event;
    const owner = repository.owner.login || repository.owner.name;
    const repo = repository.name;

    Logger.info(`收到 GitHub Push: ${owner}/${repo} ref:${ref}`);
    return { message: 'Push event received' };
  }

  /**
   * 处理 Pull Request 事件
   */
  async handlePullRequestEvent(event) {
    try {
      const { action, pull_request, repository } = event;

      // 只关注 opened, synchronize (push update), reopened
      if (!['opened', 'synchronize', 'reopened'].includes(action)) {
        return { message: `Ignored action: ${action}` };
      }

      const owner = repository.owner.login || repository.owner.name;
      const repo = repository.name;
      const prNumber = pull_request.number;
      const projectId = `${owner}/${repo}`; // GitHub Adapter 约定的 projectId 格式

      Logger.info(`处理 GitHub PR: ${projectId} #${prNumber} ${action}`);

      const request = new ReviewRequest({
        platform: 'github',
        projectId: projectId,
        mergeRequestId: prNumber,
        action: action,
        sourceBranch: pull_request.head.ref,
        targetBranch: pull_request.base.ref,
        commitSha: pull_request.head.sha,
        title: pull_request.title,
      });

      // 异步执行审查
      this.reviewService.performReview(request).catch((err) => {
        Logger.error(`GitHub PR #${prNumber} 异步审查失败`, err);
      });

      return { message: 'Review task started' };
    } catch (err) {
      Logger.error('GitHub Handler 处理失败', err);
      throw err;
    }
  }

  // 兼容性接口
  getAllTaskStatus() {
    return [];
  }
  getTaskStatus() {
    return null;
  }
}
