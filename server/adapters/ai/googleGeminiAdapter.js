import { GoogleGenerativeAI } from '@google/generative-ai';
import AIProvider from '../../core/interfaces/ai/aiProvider.js';
import Logger from '../../utils/logger.js';

class GoogleGeminiAdapter extends AIProvider {
  constructor(config) {
    super(config);
    this.genAI = new GoogleGenerativeAI(config.apiKey);
  }

  /**
   * 将 OpenAI 格式的消息转换为 Gemini 格式
   * @param {Array} messages
   */
  convertMessages(messages) {
    let systemInstruction = '';
    const history = [];
    let lastMessage = null;

    // 分离系统消息、历史消息和最新消息
    const nonSystemMessages = [];

    for (const msg of messages) {
      if (msg.role === 'system') {
        systemInstruction += (systemInstruction ? '\n' : '') + msg.content;
      } else {
        nonSystemMessages.push(msg);
      }
    }

    if (nonSystemMessages.length === 0) {
      throw new Error('No user messages found');
    }

    lastMessage = nonSystemMessages[nonSystemMessages.length - 1];
    const historyMessages = nonSystemMessages.slice(0, -1);

    for (const msg of historyMessages) {
      history.push({
        role: msg.role === 'user' ? 'user' : 'model',
        parts: [{ text: msg.content }],
      });
    }

    return { systemInstruction, history, lastMessage };
  }

  /**
   * 发送聊天完成请求
   * @param {Array<{role: string, content: string}>} messages - 消息列表
   * @param {Object} options - 选项
   * @returns {Promise<string>} AI 响应内容
   */
  async chatCompletion(messages, options = {}) {
    try {
      const { systemInstruction, history, lastMessage } = this.convertMessages(messages);

      const modelName = this.config.model || 'gemini-1.5-pro';
      const model = this.genAI.getGenerativeModel({
        model: modelName,
        systemInstruction: systemInstruction ? { parts: [{ text: systemInstruction }] } : undefined,
      });

      const generationConfig = {
        temperature: options.temperature || this.config.temperature || 0.3,
        maxOutputTokens: options.maxTokens || this.config.maxTokens || 2000,
      };

      const chat = model.startChat({
        history: history,
        generationConfig: generationConfig,
      });

      Logger.info(`Using Gemini model: ${modelName}`);

      const result = await chat.sendMessage(lastMessage.content);
      const response = await result.response;
      return response.text();
    } catch (err) {
      Logger.error('Gemini API call failed', err);
      throw err;
    }
  }
}

export default GoogleGeminiAdapter;
