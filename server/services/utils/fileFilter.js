import config from '../../config/index.js';

const AI_REVIEW_CONFIG = config.review;

class FileFilter {
  /**
   * 判断变更是否重要，需要审查
   * @param {string} diff - diff 内容
   * @param {string} fileName - 文件名
   * @returns {boolean} 是否需要审查
   */
  static isSignificantChange(diff, fileName) {
    // 检查文件类型特殊处理
    if (fileName) {
      const specialAction = this.checkSpecialFileHandling(fileName, diff);
      if (specialAction !== null) {
        return specialAction;
      }
    }

    // 检查基本变更条件
    return this.hasBasicChanges(diff);
  }

  /**
   * 检查文件类型特殊处理
   * @param {string} fileName - 文件名
   * @param {string} diff - diff 内容
   * @returns {boolean|null} 特殊处理结果，null表示无特殊处理
   */
  static checkSpecialFileHandling(fileName, diff) {
    const fileNameLower = fileName.toLowerCase();

    // 检查特殊处理文件类型
    for (const [type, conf] of Object.entries(AI_REVIEW_CONFIG.specialHandling)) {
      if (this.matchesPatterns(fileNameLower, conf.patterns)) {
        return this.handleSpecialFileType(conf.action, diff);
      }
    }

    // 检查是否在完全跳过的扩展名列表中
    if (this.matchesExtensions(fileNameLower, AI_REVIEW_CONFIG.ignoredExtensions)) {
      return false;
    }

    return null; // 无特殊处理
  }

  /**
   * 处理特殊文件类型
   * @param {string} action - 处理动作
   * @param {string} diff - diff 内容
   * @returns {boolean} 是否需要审查
   */
  static handleSpecialFileType(action, diff) {
    switch (action) {
      case 'skip':
        return false;
      case 'skipStyle':
        return this.shouldReviewVueFile(diff);
      default:
        return true;
    }
  }

  /**
   * 检查基本变更条件
   * @param {string} diff - diff 内容
   * @returns {boolean} 是否有基本变更
   */
  static hasBasicChanges(diff) {
    const addedLineCount = this.countAddedLines(diff);

    if (addedLineCount === 0) {
      return false;
    }

    return true;
  }

  /**
   * 统计新增行数
   * @param {string} diff - diff 内容
   * @returns {number} 新增行数
   */
  static countAddedLines(diff) {
    const lines = diff.split('\n');
    let addedLineCount = 0;

    for (const line of lines) {
      if (line.startsWith('+') && !line.startsWith('+++')) {
        addedLineCount++;
      }
    }

    return addedLineCount;
  }

  /**
   * 检查文件名是否匹配指定的模式
   */
  static matchesPatterns(fileNameLower, patterns) {
    return this.matchesPatternsGeneric(fileNameLower, patterns, 'includes');
  }

  /**
   * 检查文件名是否以指定的扩展名结尾
   */
  static matchesExtensions(fileNameLower, extensions) {
    return this.matchesPatternsGeneric(fileNameLower, extensions, 'endsWith');
  }

  /**
   * 通用模式匹配函数
   * @param {string} fileNameLower - 小写的文件名
   * @param {Array} patterns - 模式数组
   * @param {string} matchType - 匹配类型：'includes' 或 'endsWith'
   * @returns {boolean} 是否匹配
   */
  static matchesPatternsGeneric(fileNameLower, patterns, matchType) {
    return patterns.some((pattern) => {
      const patternLower = pattern.toLowerCase();

      switch (matchType) {
        case 'includes':
          return fileNameLower.includes(patternLower);
        case 'endsWith':
          return fileNameLower.endsWith(patternLower);
        default:
          return false;
      }
    });
  }

  /**
   * 检查Vue文件是否需要审查（跳过style部分）
   */
  static shouldReviewVueFile(diff) {
    const { addedLineCount } = this.analyzeVueFileChanges(diff);

    if (addedLineCount === 0) {
      return false;
    }

    return !this.hasOnlyStyleChanges(diff);
  }

  /**
   * 检查是否只包含style部分的变更
   */
  static hasOnlyStyleChanges(diff) {
    const { hasStyleChanges, hasNonStyleChanges } = this.analyzeVueFileChanges(diff);
    return hasStyleChanges && !hasNonStyleChanges;
  }

  /**
   * 分析Vue文件变更（公共逻辑）
   * @param {string} diff - diff 内容
   * @returns {Object} 分析结果
   */
  static analyzeVueFileChanges(diff) {
    const lines = diff.split('\n');
    let addedLineCount = 0;
    let hasStyleChanges = false;
    let hasNonStyleChanges = false;
    let inStyleSection = false;

    for (const line of lines) {
      const content = line.substring(1).trim();

      // 检查style标签边界
      if (this.isStyleTag(content)) {
        inStyleSection = !inStyleSection;
        continue;
      }

      // 统计新增行
      if (line.startsWith('+') && !line.startsWith('+++')) {
        addedLineCount++;

        if (inStyleSection) {
          hasStyleChanges = true;
        } else {
          hasNonStyleChanges = true;
        }
      }
    }

    return { addedLineCount, hasStyleChanges, hasNonStyleChanges, inStyleSection };
  }

  /**
   * 检查是否为style标签
   * @param {string} content - 内容
   * @returns {boolean} 是否为style标签
   */
  static isStyleTag(content) {
    return content.includes('<style') || content.includes('</style>');
  }
}

export default FileFilter;
