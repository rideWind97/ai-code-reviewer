import DefaultPromptStrategy from '../core/strategies/prompt/defaultStrategy';
import RepoConfigPromptStrategy from '../core/strategies/prompt/repoConfigStrategy';

class PromptService {
  /**
   * @param {import('../core/interfaces/gitPlatform')} gitPlatform
   */
  constructor(gitPlatform) {
    this.strategies = [new RepoConfigPromptStrategy(gitPlatform), new DefaultPromptStrategy()];
  }

  /**
   * 生成完整的提示词
   * @param {string} fileName
   * @param {Array} groups
   * @param {Object} context
   * @param {string} context.projectId
   * @param {string} context.ref
   */
  async generatePrompt(fileName, groups, context) {
    // 1. 构建代码上下文部分
    let prompt = `🔍 AI代码审查 - 文件: ${fileName}\n\n`;
    groups.forEach((group, index) => {
      const codeContent = group.lines.map((line) => line.content).join('\n');
      prompt += `${index + 1}. 行${group.startLine}-${group.endLine}:\n${codeContent}\n\n`;
    });

    // 2. 获取审查指令（通过策略链）
    let instructions = null;
    for (const strategy of this.strategies) {
      instructions = await strategy.getInstructions(context.projectId, context.ref);
      if (instructions) break;
    }

    if (!instructions) {
      // Fallback (should ideally be covered by DefaultStrategy)
      instructions = new DefaultPromptStrategy().getInstructions();
    }

    prompt += instructions;
    return prompt;
  }
}

export default PromptService;
