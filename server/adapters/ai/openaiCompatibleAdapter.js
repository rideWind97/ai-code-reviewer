import axios from 'axios';
import AIProvider from '../../core/interfaces/ai/aiProvider';
import Logger from '../../utils/logger';

export default class OpenAICompatibleAdapter extends AIProvider {
  constructor(config) {
    super(config);
    this.apiKey = config.apiKey;
    this.apiURL = config.apiUrl;
    this.model = config.model;
    this.timeout = config.requestTimeout || 30000;
  }

  async chatCompletion(messages, options = {}) {
    try {
      const response = await axios.post(
        `${this.apiURL}/chat/completions`,
        {
          model: this.model,
          messages: messages,
          max_tokens: options.maxTokens || this.config.maxTokens || 2000,
          temperature: options.temperature || this.config.temperature || 0.3,
          stream: false,
        },
        {
          headers: {
            Authorization: `Bearer ${this.apiKey}`,
            'Content-Type': 'application/json',
          },
          timeout: this.timeout,
        },
      );

      if (!response.data || !response.data.choices || response.data.choices.length === 0) {
        throw new Error('Invalid response from AI provider');
      }

      return response.data.choices[0].message.content.trim();
    } catch (err) {
      Logger.error('OpenAI Compatible Provider Error', err.message);
      if (err.response) {
        Logger.error('API Response Data:', err.response.data);
      }
      throw err;
    }
  }
}
