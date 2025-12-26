import PromptStrategy from '../../interfaces/promptStrategy';
import Logger from '../../../utils/logger';

class RepoConfigPromptStrategy extends PromptStrategy {
  /**
   * @param {import('../../interfaces/gitPlatform')} gitPlatform
   */
  constructor(gitPlatform) {
    super();
    this.gitPlatform = gitPlatform;
    this.configFileName = 'aiCodeReviewPrompt.cjs';
  }

  async getInstructions(projectId, ref) {
    if (!this.gitPlatform) return null;

    try {
      // 适配 GitPlatformAdapter.getProjectFile 接口 (如果不存在则需要添加)
      // 目前 GitLabAdapter 需要实现 getProjectFile
      // GitHubAdapter 也需要实现
      if (typeof this.gitPlatform.getProjectFile !== 'function') {
        return null;
      }

      const fileContent = await this.gitPlatform.getProjectFile(
        projectId,
        this.configFileName,
        ref,
      );

      if (!fileContent) return null;

      const rules = this.extractPromptRules(fileContent);
      if (rules) {
        Logger.info(`✅ 使用项目自定义审查规则: ${projectId}`);
        return rules;
      }
      return null;
    } catch (err) {
      Logger.warn(`获取项目自定义规则失败: ${projectId}`, err.message);
      return null;
    }
  }

  extractPromptRules(fileContent) {
    try {
      // 1. 匹配 static getProjectPrompt() 方法体
      const getProjectPromptMatch = fileContent.match(
        /static\s+getProjectPrompt\s*\(\s*\)\s*\{[\s\S]*?return\s+prompt;\s*\}/,
      );

      if (getProjectPromptMatch) {
        const methodContent = getProjectPromptMatch[0];

        // 尝试匹配连续的 prompt += `...`
        const promptMatches = methodContent.match(/prompt\s*\+=\s*`([^`]+)`/g);
        if (promptMatches && promptMatches.length > 0) {
          return promptMatches.map((m) => m.match(/prompt\s*\+=\s*`([^`]+)`/)[1]).join('\n');
        }

        // 尝试匹配 let prompt = `...`
        const simpleMatch = methodContent.match(/let\s+prompt\s*=\s*`([\s\S]*?)`;/);
        if (simpleMatch) return simpleMatch[1];
      }

      // 2. 兼容旧的简单模式匹配
      const simpleMatches = fileContent.match(/prompt\s*\+=\s*`([^`]+)`/g);
      if (simpleMatches) {
        let rules = '请按照以下规则进行专业代码审查:\n\n';
        rules += simpleMatches.map((m) => m.match(/prompt\s*\+=\s*`([^`]+)`/)[1]).join('\n');
        return rules;
      }

      return null;
    } catch (err) {
      Logger.error('解析项目Prompt文件失败', err);
      return null;
    }
  }
}

export default RepoConfigPromptStrategy;
