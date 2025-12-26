/**
 * Git 平台适配器接口
 * 所有具体的平台实现（GitLab, GitHub）都必须继承此类
 */
export default class GitPlatformAdapter {
  constructor(config) {
    this.config = config;
  }

  /**
   * 获取代码变更
   * @param {import('../models/reviewRequest')} reviewRequest
   * @returns {Promise<import('../models/codeChange')[]>}
   */
  async getChanges(reviewRequest) {
    throw new Error('getChanges must be implemented');
  }

  /**
   * 获取已有的评论
   * @param {import('../models/reviewRequest')} reviewRequest
   * @returns {Promise<import('../models/reviewComment')[]>}
   */
  async getExistingComments(reviewRequest) {
    throw new Error('getExistingComments must be implemented');
  }

  /**
   * 发布评论
   * @param {import('../models/reviewRequest')} reviewRequest
   * @param {import('../models/reviewComment')[]} comments
   * @returns {Promise<Object>} 发布结果统计
   */
  async publishComments(reviewRequest, comments) {
    throw new Error('publishComments must be implemented');
  }

  /**
   * 获取项目文件内容
   * @param {string|number} projectId
   * @param {string} filePath
   * @param {string} ref
   * @returns {Promise<string|null>}
   */
  async getProjectFile(projectId, filePath, ref) {
    throw new Error('getProjectFile must be implemented');
  }
}
