import Logger from '../utils/logger';
import config from '../config/index.js';
import GitLabAdapter from '../adapters/gitlabAdapter';
import CodeReviewService from '../services/codeReviewService';
import AIApiService from '../services/aiApiService';
import PromptService from '../services/promptService';
import ReviewRequest from '../core/models/reviewRequest';

/**
 * GitLab Webhook 处理器 - 新架构适配版
 */
export default class GitlabEventHandler {
  constructor() {
    // 依赖组装
    this.config = config;
    this.gitAdapter = new GitLabAdapter(config);
    this.promptService = new PromptService(this.gitAdapter);
    this.aiService = new AIApiService(config.ai, this.promptService);
    this.reviewService = new CodeReviewService(this.gitAdapter, this.aiService, config);
  }

  /**
   * 处理推送事件
   */
  async handlePushEvent(event) {
    Logger.info(`收到 Push 事件: ${event.project_id}`);
    return { message: 'Push event received, no action needed' };
  }

  /**
   * 处理合并请求事件
   */
  async handleMergeRequestEvent(event) {
    try {
      const { project, object_attributes } = event;
      const { iid, action, state, last_commit } = object_attributes;
      const projectId = project.id;

      Logger.info(`处理 MR 事件: 项目 ${projectId}, MR ${iid}, 动作 ${action}`);

      if (action === 'open' || action === 'update') {
        // 构建标准请求对象
        const request = new ReviewRequest({
          platform: 'gitlab',
          projectId: projectId,
          mergeRequestId: iid,
          action: action,
          sourceBranch: object_attributes.source_branch,
          targetBranch: object_attributes.target_branch,
          commitSha: last_commit && last_commit.id ? last_commit.id : null,
          title: object_attributes.title,
        });

        // 异步执行审查
        this.reviewService.performReview(request).catch((err) => {
          Logger.error(`异步审查失败 MR ${iid}`, err);
        });

        return { message: 'Review task started' };
      }

      return { message: `MR ${action} ignored` };
    } catch (err) {
      Logger.error('Handler 处理失败', err);
      throw err;
    }
  }

  // 兼容旧接口
  getAllTaskStatus() {
    return [];
  }
  getTaskStatus() {
    return null;
  }
  getProjectTaskStatus() {
    return [];
  }
}
