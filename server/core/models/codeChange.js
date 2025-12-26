/**
 * @typedef {Object} CodeChangeData
 * @property {string} filePath - 文件路径 (优先取 new_path)
 * @property {string} [oldPath] - 旧文件路径
 * @property {string} content - 变更内容 (diff)
 * @property {boolean} isNewFile - 是否为新文件
 * @property {boolean} isDeleted - 是否已删除
 * @property {string} [baseSha] - 基础 SHA
 * @property {string} [headSha] - 头部 SHA
 */

export default class CodeChange {
  /**
   * @param {CodeChangeData} data
   */
  constructor(data) {
    this.filePath = data.filePath;
    this.oldPath = data.oldPath || null;
    this.content = data.content;
    this.isNewFile = !!data.isNewFile;
    this.isDeleted = !!data.isDeleted;
    this.baseSha = data.baseSha || null;
    this.headSha = data.headSha || null;
  }
}
