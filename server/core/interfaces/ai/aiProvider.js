/**
 * AI Provider 接口
 */
export default class AIProvider {
  constructor(config) {
    this.config = config;
  }

  /**
   * 发送聊天完成请求
   * @param {Array<{role: string, content: string}>} messages - 消息列表
   * @param {Object} options - 选项
   * @param {number} [options.temperature]
   * @param {number} [options.maxTokens]
   * @returns {Promise<string>} AI 响应内容
   */
  async chatCompletion(messages, options = {}) {
    throw new Error('chatCompletion must be implemented');
  }
}
