import Logger from '../utils/logger';
import OpenAICompatibleAdapter from '../adapters/ai/openaiCompatibleAdapter';
import GoogleGeminiAdapter from '../adapters/ai/googleGeminiAdapter';
import config from '../config/index.js';

const REVIEW_CONFIG = config.review;

class AIApiService {
  /**
   * @param {Object} aiConfig - AI 配置对象
   * @param {import('./promptService')} promptService
   */
  constructor(aiConfig, promptService) {
    this.config = aiConfig;
    this.promptService = promptService;
    this.provider = this.createProvider(aiConfig);
  }

  createProvider(config) {
    switch (config.provider) {
      case 'openai':
      case 'deepseek': // deepseek is openai compatible
        return new OpenAICompatibleAdapter(config);
      case 'gemini':
        return new GoogleGeminiAdapter(config);
      // case 'anthropic':
      //   return new AnthropicAdapter(config);
      default:
        Logger.warn(`Unknown AI provider '${config.provider}', defaulting to OpenAI compatible`);
        return new OpenAICompatibleAdapter(config);
    }
  }

  /**
   * 批量生成审查意见
   * @param {string} fileName - 文件名
   * @param {Array} groups - 代码组数组
   * @param {Array} existingComments - 已有评论数组
   * @param {Object} context - 上下文 (projectId, ref)
   * @returns {Promise<Array>} 审查结果数组
   */
  async generateBatchReview(fileName, groups, existingComments, context = {}) {
    try {
      const reviewText = await this.callAIForReview(fileName, groups, context);
      const results = this.parseBatchReviewResultFast(fileName, groups, reviewText);

      Logger.info(`📊 批量审查解析结果 [${fileName}]:`, {
        totalGroups: groups.length,
        generatedReviews: results.length,
        results: results.map((r) => ({
          lineNumber: r.lineNumber,
          reviewPreview: r.review.substring(0, 100) + '...',
        })),
      });

      return results;
    } catch (err) {
      Logger.error('批量审查失败', err);
      return [];
    }
  }

  /**
   * 调用AI进行代码审查（私有方法）
   */
  async callAIForReview(fileName, groups, context = {}) {
    // 使用 PromptService 生成 Prompt
    let prompt;
    if (this.promptService) {
      prompt = await this.promptService.generatePrompt(fileName, groups, context);
    } else {
      Logger.warn('PromptService not initialized, using basic prompt');
      prompt = `Review this code:\n\n` + JSON.stringify(groups);
    }

    const messages = [
      {
        role: 'system',
        content:
          "你是代码审查专家。分析代码组，只为有问题的组提供具体的改进建议。如果代码完全没有问题，请直接回复'PASS'，不要生成任何评论。用中文，简洁。",
      },
      {
        role: 'user',
        content: prompt,
      },
    ];

    return await this.provider.chatCompletion(messages);
  }

  /**
   * 快速解析批量审查结果
   */
  parseBatchReviewResultFast(fileName, groups, reviewText) {
    const reviews = [];

    if (!reviewText) return [];

    try {
      // 检查是否所有组都通过审查
      const lines = reviewText.split('\n').filter((line) => line.trim());
      const allPass = lines.every((line) => line.includes('PASS'));

      if (allPass) {
        return reviews;
      }

      if (reviewText.trim() && reviewText.length > 10) {
        // 解析带编号的响应 - 确保处理所有组
        lines.forEach((line, index) => {
          if (line.includes('PASS')) {
            return; // 跳过PASS的组
          }

          const group = groups[index];
          if (group) {
            const lastLine = group.lines[group.lines.length - 1];

            // 清理响应文本，移除编号
            const cleanReview = line.replace(/^\d+\.\s*/, '').trim();

            if (cleanReview && cleanReview !== 'PASS') {
              reviews.push({
                lineNumber: lastLine.lineNumber,
                content: lastLine.content,
                review: cleanReview,
                groupId: group.id,
                isGroupEnd: true,
                groupSize: group.lines.length,
              });
            }
          }
        });

        // 如果没有解析到任何结果，尝试其他解析方式
        if (reviews.length === 0) {
          // 尝试按组数量分割响应
          const groupCount = groups.length;
          const responseLines = reviewText.split('\n').filter((line) => line.trim());

          if (responseLines.length >= groupCount) {
            // 如果响应行数大于等于组数，按组分配
            responseLines.slice(0, groupCount).forEach((line, index) => {
              const group = groups[index];
              if (group && !line.includes('PASS')) {
                const lastLine = group.lines[group.lines.length - 1];
                const cleanReview = line.replace(/^\d+\.\s*/, '').trim();

                if (cleanReview && cleanReview !== 'PASS') {
                  reviews.push({
                    lineNumber: lastLine.lineNumber,
                    content: lastLine.content,
                    review: cleanReview,
                    groupId: group.id,
                    isGroupEnd: true,
                    groupSize: group.lines.length,
                  });
                }
              }
            });
          } else {
            // 如果响应行数少于组数，将响应分配给第一个组
            const firstGroup = groups[0];
            if (firstGroup) {
              const lastLine = firstGroup.lines[firstGroup.lines.length - 1];
              reviews.push({
                lineNumber: lastLine.lineNumber,
                content: lastLine.content,
                review: reviewText.trim(),
                groupId: firstGroup.id,
                isGroupEnd: true,
                groupSize: firstGroup.lines.length,
              });
            }
          }
        }
      }
    } catch (err) {
      Logger.error('快速解析批量审查结果失败', err);
    }

    return reviews;
  }

  /**
   * 判断审查意见是否质量较低
   */
  isLowQualityReview(review) {
    return REVIEW_CONFIG.lowQualityPatterns.some((pattern) =>
      review.toLowerCase().includes(pattern.toLowerCase()),
    );
  }
}

export default AIApiService;
