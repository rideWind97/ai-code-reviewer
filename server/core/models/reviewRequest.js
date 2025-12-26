/**
 * @typedef {Object} ReviewRequestData
 * @property {string} platform - 平台名称 (gitlab, github)
 * @property {string|number} projectId - 项目ID或仓库名
 * @property {string|number} mergeRequestId - MR/PR 编号
 * @property {string} action - 触发动作 (open, update, push)
 * @property {string} [sourceBranch] - 源分支
 * @property {string} [targetBranch] - 目标分支
 * @property {string} [commitSha] - 关联的 Commit SHA
 * @property {string} [title] - MR/PR 标题
 */

export default class ReviewRequest {
  /**
   * @param {ReviewRequestData} data
   */
  constructor(data) {
    this.platform = data.platform;
    this.projectId = data.projectId;
    this.mergeRequestId = data.mergeRequestId;
    this.action = data.action;
    this.sourceBranch = data.sourceBranch;
    this.targetBranch = data.targetBranch;
    this.commitSha = data.commitSha;
    this.title = data.title || '';
    this.createdAt = new Date();
  }
}
