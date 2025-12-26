import Logger from '../utils/logger';
import CodeProcessor from './utils/codeProcessor';
import ReviewComment from '../core/models/reviewComment';
import FileFilter from './utils/fileFilter';

class CodeReviewService {
  /**
   * @param {import('../core/interfaces/gitPlatform')} gitPlatform
   * @param {import('./aiApiService')} aiService
   * @param {Object} config
   */
  constructor(gitPlatform, aiService, config) {
    this.gitPlatform = gitPlatform;
    this.aiService = aiService;
    this.config = config;
  }

  /**
   * 执行代码审查的主流程
   * @param {import('../core/models/reviewRequest')} reviewRequest
   */
  async performReview(reviewRequest) {
    const startTime = Logger.startTimer('CodeReview');

    try {
      // 0. 检查是否需要跳过 (基于标题)
      if (this.shouldSkipReview(reviewRequest)) {
        Logger.info(`审查被跳过: ${reviewRequest.title}`);
        return;
      }

      // 1. 获取代码变更
      const changes = await this.gitPlatform.getChanges(reviewRequest);
      if (!changes || changes.length === 0) {
        Logger.info('没有代码变更，跳过审查');
        return;
      }

      // 2. 获取已有评论（用于增量审查去重）
      const existingComments = await this.gitPlatform.getExistingComments(reviewRequest);

      // 3. 预处理和过滤
      const significantChanges = changes.filter((change) => this.isSignificantChange(change));

      if (significantChanges.length === 0) {
        Logger.info('没有需要审查的重要变更');
        return;
      }

      // 4. 并发执行审查
      const newComments = await this.processFilesConcurrently(
        significantChanges,
        existingComments,
        reviewRequest,
      );

      // 5. 发布评论
      if (newComments.length > 0) {
        await this.gitPlatform.publishComments(reviewRequest, newComments);
      }

      Logger.endTimer('CodeReview', startTime, {
        files: changes.length,
        newComments: newComments.length,
      });
    } catch (err) {
      Logger.error('代码审查流程失败', err);
      throw err;
    }
  }

  shouldSkipReview(reviewRequest) {
    const { ignoreCrKeyword } = this.config.review;
    if (
      reviewRequest.title &&
      ignoreCrKeyword &&
      reviewRequest.title.toLowerCase().includes(ignoreCrKeyword.toLowerCase())
    ) {
      return true;
    }
    return false;
  }

  isSignificantChange(change) {
    // 1. 过滤掉删除的文件
    if (change.isDeleted) return false;

    // 2. 复用 FileFilter 的逻辑
    return FileFilter.isSignificantChange(change.content, change.filePath);
  }

  async processFilesConcurrently(changes, existingComments, reviewRequest) {
    const allComments = [];
    const { maxConcurrentFiles } = this.config.review;

    // 简单的分批处理
    for (let i = 0; i < changes.length; i += maxConcurrentFiles) {
      const batch = changes.slice(i, i + maxConcurrentFiles);
      const batchPromises = batch.map((change) =>
        this.reviewSingleFile(change, existingComments, reviewRequest),
      );

      const results = await Promise.allSettled(batchPromises);
      results.forEach((res) => {
        if (res.status === 'fulfilled' && res.value) {
          allComments.push(...res.value);
        }
      });
    }
    return allComments;
  }

  async reviewSingleFile(change, existingComments, reviewRequest) {
    try {
      // 解析 Diff
      const changeLines = CodeProcessor.parseDiffLines(change.content);
      if (changeLines.length === 0) return [];

      // 预过滤代码行
      const filteredLines = CodeProcessor.preFilterCodeLines(changeLines);
      if (filteredLines.length === 0) return [];

      // 分组
      const groups = CodeProcessor.optimizeLineGroups(filteredLines);
      if (groups.length === 0) return [];

      // 过滤掉已有评论的组
      const groupsToReview = groups.filter(
        (group) => !this.hasComment(group, change.filePath, existingComments),
      );

      if (groupsToReview.length === 0) return [];

      // 调用 AI
      const options = {
        projectId: reviewRequest.projectId,
        ref: reviewRequest.targetBranch || 'main',
        gitlabAPI: null,
      };

      if (this.gitPlatform.constructor.name === 'GitLabAdapter') {
        options.gitlabAPI = this.gitPlatform;
      }

      const aiResults = await this.aiService.generateBatchReview(
        change.filePath,
        groupsToReview,
        existingComments,
        options,
      );

      // 转换结果为 ReviewComment 模型
      return aiResults.map(
        (res) =>
          new ReviewComment({
            filePath: change.filePath,
            lineNumber: res.lineNumber,
            content: res.review,
            side: 'RIGHT',
          }),
      );
    } catch (err) {
      Logger.error(`审查文件失败 ${change.filePath}`, err);
      return [];
    }
  }

  hasComment(group, filePath, existingComments) {
    return existingComments.some(
      (c) =>
        c.filePath === filePath && c.lineNumber >= group.startLine && c.lineNumber <= group.endLine,
    );
  }
}

export default CodeReviewService;
