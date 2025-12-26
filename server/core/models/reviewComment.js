/**
 * @typedef {Object} ReviewCommentData
 * @property {string} filePath - 文件路径
 * @property {number} lineNumber - 行号
 * @property {string} content - 评论内容
 * @property {string} [originalContent] - 被评论的代码内容
 * @property {string} [side] - 评论位置 (LEFT, RIGHT)
 */

export default class ReviewComment {
  /**
   * @param {ReviewCommentData} data
   */
  constructor(data) {
    this.filePath = data.filePath;
    this.lineNumber = data.lineNumber;
    this.content = data.content;
    this.originalContent = data.originalContent || null;
    this.side = data.side || 'RIGHT'; // 默认在右侧（新增代码侧）
    this.createdAt = new Date();
  }
}
