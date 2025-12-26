/**
 * Prompt 生成策略接口
 */
export default class PromptStrategy {
  /**
   * 获取系统提示词/审查规则
   * @param {string} _projectId - 项目ID
   * @param {string} _ref - Git引用
   * @returns {Promise<string|null>} 提示词规则，如果策略不适用返回 null
   */
  async getInstructions(_projectId, _ref) {
    throw new Error('Not implemented');
  }
}
