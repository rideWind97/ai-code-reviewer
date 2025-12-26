import axios from 'axios';
import Logger from '../../utils/logger';
import config from '../../config/index.js';

const { ignoreCr } = config.review;

/**
 * 极简化 GitLab 代码审查服务
 */
class SimpleGitlabCR {
  constructor() {
    this.baseURL = config.gitlab.url;
    this.token = config.gitlab.token;

    // 配置常量
    this.config = {
      timeout: 10000,
      perPage: 100,
      maxFiles: 100,
      apiDelay: 200,
      rateLimitDelay: 100,
    };

    // API 端点常量
    this.endpoints = {
      mr: (projectId, mrIid) => `/api/v4/projects/${projectId}/merge_requests/${mrIid}`,
      mrChanges: (projectId, mrIid) =>
        `/api/v4/projects/${projectId}/merge_requests/${mrIid}/changes`,
      mrNotes: (projectId, mrIid, page, perPage) =>
        `/api/v4/projects/${projectId}/merge_requests/${mrIid}/notes?page=${page}&per_page=${perPage}`,
      mrDiscussions: (projectId, mrIid) =>
        `/api/v4/projects/${projectId}/merge_requests/${mrIid}/discussions`,
      commit: (projectId, sha) => `/api/v4/projects/${projectId}/repository/commits/${sha}`,
      commitDiff: (projectId, sha) =>
        `/api/v4/projects/${projectId}/repository/commits/${sha}/diff`,
      repositoryFile: (projectId, filePath, ref = 'main') =>
        `/api/v4/projects/${projectId}/repository/files/${encodeURIComponent(filePath)}/raw?ref=${ref}`,
    };
  }

  /**
   * 通用API调用方法
   * @param {string} url - API URL
   * @param {string} description - 操作描述
   * @returns {Promise<Object>} API响应数据
   */
  async makeAPICall(url, description) {
    try {
      const response = await axios.get(url, {
        headers: { 'PRIVATE-TOKEN': this.token },
        timeout: this.config.timeout,
      });
      return response.data;
    } catch (err) {
      Logger.error(`${description}失败:`, err.message);
      throw err;
    }
  }

  /**
   * 获取 MR 变更内容
   * @param {string} projectId - 项目ID
   * @param {string} mrIid - MR IID
   * @param {string} action - MR动作 (open/update)
   * @param {string} lastCommitSha - 最后一次commit的SHA (仅在update时使用)
   */
  async getMRChanges(projectId, mrIid, action = 'open', lastCommitSha = null) {
    try {
      // 首先获取 MR 详细信息
      const mrInfo = await this.makeAPICall(
        `${this.baseURL}${this.endpoints.mr(projectId, mrIid)}`,
        '获取 MR 详细信息',
      );

      if (mrInfo.title && mrInfo.title.toLowerCase().includes(ignoreCr)) {
        Logger.info(`🚫 MR标题包含"no-cr"，跳过代码审查: ${mrInfo.title}`);
        return { skipReview: true, title: mrInfo.title };
      }

      let changes = [];

      if (action === 'update' && lastCommitSha) {
        Logger.info(`🔄 MR更新模式：检查文件数量决定审查策略`);
        const allChangesData = await this.makeAPICall(
          `${this.baseURL}${this.endpoints.mrChanges(projectId, mrIid)}`,
          '获取所有MR变更',
        );
        const allChanges = allChangesData.changes || [];

        if (allChanges.length < 20) {
          Logger.info(`📊 文件数量 ${allChanges.length} < 20，执行全量CR`);
          changes = allChanges;
        } else {
          Logger.info(
            `📊 文件数量 ${allChanges.length} >= 20，只审查最后一次commit ${lastCommitSha}`,
          );
          changes = await this.getLastCommitChanges(projectId, mrIid, lastCommitSha, mrInfo);
        }
      } else {
        // MR打开时，获取所有变更
        Logger.info(`🆕 MR打开模式：审查所有变更`);
        const changesData = await this.makeAPICall(
          `${this.baseURL}${this.endpoints.mrChanges(projectId, mrIid)}`,
          '获取 MR 变更',
        );
        changes = changesData.changes || [];
      }

      // 为每个变更添加必要的 SHA 信息
      const enrichedChanges = this.enrichChangesWithSHA(changes, mrInfo);

      Logger.info(`📁 获取到 ${enrichedChanges.length} 个文件变更`);
      return enrichedChanges;
    } catch (err) {
      Logger.error('获取 MR 变更失败:', err.message);
      return [];
    }
  }

  /**
   * 为变更数据添加SHA信息
   * @param {Array} changes - 变更数组
   * @param {Object} mrInfo - MR信息
   * @returns {Array} 增强后的变更数组
   */
  enrichChangesWithSHA(changes, mrInfo) {
    return changes.map((change) => ({
      ...change,
      base_sha: mrInfo.diff_refs?.base_sha || mrInfo.sha,
      start_sha: mrInfo.diff_refs?.start_sha || mrInfo.sha,
      head_sha: mrInfo.diff_refs?.head_sha || mrInfo.sha,
    }));
  }

  /**
   * 将commit diff转换为MR changes格式
   * @param {Array} diffData - diff数据
   * @param {Object} commitInfo - commit信息
   * @param {string} lastCommitSha - commit SHA
   * @returns {Array} 转换后的变更数组
   */
  convertCommitDiffToChanges(diffData, commitInfo, lastCommitSha) {
    return diffData.map((diff) => ({
      old_path: diff.old_path,
      new_path: diff.new_path,
      a_mode: diff.a_mode,
      b_mode: diff.b_mode,
      diff: diff.diff,
      new_file: diff.new_file,
      renamed_file: diff.renamed_file,
      deleted_file: diff.deleted_file,
      // 添加commit相关信息
      commit_sha: lastCommitSha,
      commit_message: commitInfo.message,
      commit_author: commitInfo.author_name,
      commit_date: commitInfo.created_at,
    }));
  }

  /**
   * 获取最后一次commit的变更
   * @param {string} projectId - 项目ID
   * @param {string} mrIid - MR IID
   * @param {string} lastCommitSha - 最后一次commit的SHA
   * @param {Object} mrInfo - MR信息
   */
  async getLastCommitChanges(projectId, mrIid, lastCommitSha, mrInfo) {
    try {
      // 获取最后一次commit的详细信息
      const commitInfo = await this.makeAPICall(
        `${this.baseURL}${this.endpoints.commit(projectId, lastCommitSha)}`,
        '获取commit详细信息',
      );

      // 获取该commit的diff
      const diffData = await this.makeAPICall(
        `${this.baseURL}${this.endpoints.commitDiff(projectId, lastCommitSha)}`,
        '获取commit diff',
      );

      // 将diff数据转换为与MR changes相同的格式
      const changes = this.convertCommitDiffToChanges(diffData, commitInfo, lastCommitSha);

      Logger.info(`📝 最后一次commit包含 ${changes.length} 个文件变更`);
      return changes;
    } catch (err) {
      Logger.error('获取最后一次commit变更失败:', err.message);
      Logger.warn('回退到获取所有MR变更');
      const changesData = await this.makeAPICall(
        `${this.baseURL}${this.endpoints.mrChanges(projectId, mrIid)}`,
        '获取所有MR变更',
      );
      return changesData.changes || [];
    }
  }

  /**
   * 延迟执行
   * @param {number} ms - 延迟毫秒数
   */
  async delay(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  /**
   * 解析评论数据
   * @param {Array} allNotes - 所有notes数据
   * @returns {Array} 解析后的评论数组
   */
  parseComments(allNotes) {
    const comments = [];

    for (const note of allNotes) {
      // 检查是否有位置信息（行内评论）
      if (note.position && note.position.new_line) {
        comments.push({
          filePath: note.position.new_path,
          line: note.position.new_line,
          startLine: note.position.new_line,
          endLine: note.position.new_line,
          note: note.body,
          noteId: note.id,
          createdAt: note.created_at,
        });
      } else if (note.body && note.body.includes('AI Review')) {
        // 检查是否是 AI 生成的评论
        comments.push({
          filePath: 'general',
          line: null,
          startLine: null,
          endLine: null,
          note: note.body,
          noteId: note.id,
          createdAt: note.created_at,
          isAIGenerated: true,
        });
      }
    }

    return comments;
  }

  /**
   * 获取已有评论（支持分页）
   */
  async getExistingComments(projectId, mrIid) {
    try {
      let allNotes = [];
      let page = 1;

      // 分页获取所有评论
      while (true) {
        const notes = await this.makeAPICall(
          `${this.baseURL}${this.endpoints.mrNotes(projectId, mrIid, page, this.config.perPage)}`,
          `获取第${page}页评论`,
        );

        allNotes = allNotes.concat(notes);

        // 如果返回的评论数少于每页数量，说明已经是最后一页
        if (notes.length < this.config.perPage) {
          break;
        }

        page++;

        // 避免 API 频率限制
        await this.delay(this.config.rateLimitDelay);
      }

      Logger.info(`🔍 获取到 ${allNotes.length} 个 notes，开始解析...`);

      const comments = this.parseComments(allNotes);

      Logger.info(`✅ 成功解析出 ${comments.length} 个已有评论`);
      return comments;
    } catch (err) {
      Logger.error('获取已有评论失败:', err.message);
      return [];
    }
  }

  /**
   * 检查行是否已有评论
   */
  hasCommentOnLine(existingComments, filePath, lineNumber) {
    return existingComments.some(
      (comment) => comment.filePath === filePath && comment.line === lineNumber,
    );
  }

  /**
   * 检查是否已有相似的评论（更严格的检查）
   */
  hasSimilarCommentOnLine(existingComments, filePath, lineNumber, reviewText) {
    return existingComments.some((comment) => {
      if (comment.filePath !== filePath || comment.line !== lineNumber) {
        return false;
      }

      // 检查评论内容是否相似（简单的相似度检查）
      const existingText = comment.note || comment.body || '';
      const similarity = this.calculateTextSimilarity(existingText, reviewText);

      // 如果相似度超过70%，认为是重复评论
      return similarity > 0.7;
    });
  }

  /**
   * 计算两个文本的相似度
   */
  calculateTextSimilarity(text1, text2) {
    const words1 = text1.toLowerCase().split(/\s+/);
    const words2 = text2.toLowerCase().split(/\s+/);

    const set1 = new Set(words1);
    const set2 = new Set(words2);

    const intersection = new Set([...set1].filter((x) => set2.has(x)));
    const union = new Set([...set1, ...set2]);

    return intersection.size / union.size;
  }

  /**
   * 通用POST API调用方法
   * @param {string} url - API URL
   * @param {Object} data - 请求数据
   * @param {string} description - 操作描述
   * @returns {Promise<Object>} API响应数据
   */
  async makePostAPICall(url, data, description) {
    try {
      const response = await axios.post(url, data, {
        headers: { 'PRIVATE-TOKEN': this.token },
        timeout: this.config.timeout,
      });
      return response.data;
    } catch (err) {
      Logger.error(`${description}失败:`, err.message);
      if (err.response) {
        Logger.error('API 响应状态:', err.response.status);
        Logger.error('API 响应数据:', err.response.data);
      }
      throw err;
    }
  }

  /**
   * 创建评论位置对象
   * @param {Object} change - 变更对象
   * @param {number} lineNumber - 行号
   * @returns {Object} 位置对象
   */
  createCommentPosition(change, lineNumber) {
    return {
      base_sha: change.base_sha,
      start_sha: change.start_sha,
      head_sha: change.head_sha,
      old_path: change.old_path,
      new_path: change.new_path,
      position_type: 'text',
      old_line: null,
      new_line: lineNumber,
    };
  }

  /**
   * 发布行内评论
   */
  async postInlineComment(projectId, mrIid, filePath, lineNumber, comment, change) {
    try {
      const position = this.createCommentPosition(change, lineNumber);

      Logger.info(`🔍 发布行内评论: ${filePath}:${lineNumber}`, { position });

      await this.makePostAPICall(
        `${this.baseURL}${this.endpoints.mrDiscussions(projectId, mrIid)}`,
        {
          body: comment,
          position: position,
        },
        `发布行内评论 ${filePath}:${lineNumber}`,
      );

      Logger.info(`✅ 评论发布成功: ${filePath}:${lineNumber}`);
      return true;
    } catch (err) {
      Logger.error(`❌ 评论发布失败: ${filePath}:${lineNumber}`, err.message);
      return false;
    }
  }

  /**
   * 执行代码审查（支持大量文件）
   * @param {string} projectId - 项目ID
   * @param {string} mrIid - MR IID
   * @param {Array} fileReviews - 文件审查结果
   * @param {string} action - MR动作 (open/update)
   * @param {string} lastCommitSha - 最后一次commit的SHA (仅在update时使用)
   */
  async executeCodeReview(projectId, mrIid, fileReviews, action = 'open', lastCommitSha = null) {
    try {
      // 获取变更和已有评论
      const changes = await this.getMRChanges(projectId, mrIid, action, lastCommitSha);

      // 检查是否需要跳过代码审查
      if (changes && changes.skipReview) {
        Logger.info(`🚫 跳过代码审查: ${changes.title}`);
        return {
          successCount: 0,
          skippedCount: 0,
          totalProcessed: 0,
          filesProcessed: 0,
          skipped: true,
          reason: 'MR标题包含"no-cr"',
        };
      }

      const existingComments = await this.getExistingComments(projectId, mrIid);

      if (!changes.length) {
        Logger.info('没有代码变更，跳过审查');
        return;
      }

      // 文件数量限制和性能优化
      if (fileReviews.length > this.config.maxFiles) {
        Logger.warn(
          `⚠️ 文件数量过多 (${fileReviews.length})，限制处理前 ${this.config.maxFiles} 个文件`,
        );
        fileReviews = fileReviews.slice(0, this.config.maxFiles);
      }

      let successCount = 0;
      let skipCount = 0;
      let totalProcessed = 0;

      Logger.info(`开始处理 ${fileReviews.length} 个文件的代码审查`);

      // 遍历每个文件的审查结果
      for (const fileReview of fileReviews) {
        const filePath = fileReview.filePath;

        // 找到对应的变更获取 SHA
        const change = changes.find((c) => (c.new_path || c.old_path) === filePath);
        if (!change) {
          Logger.warn(`⚠️ 未找到文件 ${filePath} 对应的变更信息`);
          continue;
        }

        let fileCommentCount = 0;

        // 处理每行评论
        for (const lineReview of fileReview.review) {
          if (!lineReview.isGroupEnd) continue;
          const lineNumber = lineReview.lineNumber;

          // 检查是否已有评论（更严格的检查）
          if (
            this.hasCommentOnLine(existingComments, filePath, lineNumber) ||
            this.hasSimilarCommentOnLine(existingComments, filePath, lineNumber, lineReview.review)
          ) {
            Logger.info(`⚠️ 跳过重复评论: ${filePath}:${lineNumber}`);
            skipCount++;
            continue;
          }

          // 发布评论
          const success = await this.postInlineComment(
            projectId,
            mrIid,
            filePath,
            lineNumber,
            lineReview.review,
            change,
          );

          if (success) {
            successCount++;
            fileCommentCount++;
            totalProcessed++;
          }

          // 避免 API 频率限制
          await this.delay(this.config.apiDelay);
        }

        Logger.info(`📁 文件 ${filePath} 处理完成: ${fileCommentCount} 个评论`);
      }

      Logger.info(
        `代码审查完成: 成功 ${successCount} 个，跳过 ${skipCount} 个，总计处理 ${totalProcessed} 个评论`,
      );

      return {
        successCount,
        skippedCount: skipCount,
        totalProcessed,
        filesProcessed: fileReviews.length,
      };
    } catch (err) {
      Logger.error('代码审查执行失败:', err.message);
      throw err;
    }
  }

  /**
   * 获取项目文件内容
   * @param {string} projectId - 项目ID
   * @param {string} filePath - 文件路径
   * @param {string} ref - 分支或commit SHA，默认为 'main'
   * @returns {Promise<string|null>} 文件内容，如果文件不存在返回null
   */
  async getProjectFile(projectId, filePath, ref = 'main') {
    try {
      const url = `${this.baseURL}${this.endpoints.repositoryFile(projectId, filePath, ref)}`;
      const response = await axios.get(url, {
        headers: { 'PRIVATE-TOKEN': this.token },
        timeout: this.config.timeout,
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

export default SimpleGitlabCR;
