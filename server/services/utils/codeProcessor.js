import config from '../../config/index.js';

const AI_REVIEW_CONFIG = config.review;

class CodeProcessor {
  /**
   * 智能预过滤代码行
   * @param {Array} changeLines - 变更行数组
   * @returns {Array} 过滤后的变更行数组
   */
  static preFilterCodeLines(changeLines) {
    return changeLines.filter((line) => {
      const content = line.content.trim();

      // 长度检查
      if (!this.isValidContentLength(content)) {
        return false;
      }

      // 跳过模式检查
      if (this.matchesSkipPatterns(content)) {
        return false;
      }

      // 空白内容检查
      // if (!this.isNotJustWhitespace(content)) {
      //   return false;
      // }

      return true;
    });
  }

  /**
   * 检查内容长度是否有效
   * @param {string} content - 内容
   * @returns {boolean} 是否有效
   */
  static isValidContentLength(content) {
    // Hardcoded defaults if quickCheckRules missing from new config
    const minLength = 3;
    const maxLength = 200;

    if (content.length < minLength) {
      return false;
    }

    if (content.length > maxLength) {
      return true; // 内容过长，直接保留无需进一步检查
    }

    return true;
  }

  /**
   * 检查是否匹配跳过模式
   * @param {string} content - 内容
   * @returns {boolean} 是否匹配跳过模式
   */
  static matchesSkipPatterns(content) {
    return AI_REVIEW_CONFIG.skipPatterns.some((pattern) => pattern.test(content));
  }

  /**
   * 优化代码行分组策略
   * @param {Array} changeLines - 变更行数组
   * @returns {Array} 优化后的分组数组
   */
  static optimizeLineGroups(changeLines) {
    return this.groupLines(changeLines, true);
  }

  /**
   * 通用的代码行分组方法
   * @param {Array} changeLines - 变更行数组
   * @param {boolean} useSmartGrouping - 是否使用智能分组
   * @returns {Array} 分组后的数组
   */
  static groupLines(changeLines, useSmartGrouping = false) {
    if (changeLines.length === 0) return [];

    const groups = [];
    let currentGroup = this.createNewGroup(changeLines[0], 1);

    for (let i = 1; i < changeLines.length; i++) {
      const currentLine = changeLines[i];
      const lastLine = currentGroup.lines[currentGroup.lines.length - 1];

      const shouldStartNewGroup = useSmartGrouping
        ? this.shouldStartNewGroup(currentLine, lastLine, currentGroup)
        : this.shouldStartNewGroupBasic(currentLine, lastLine, currentGroup);

      if (shouldStartNewGroup) {
        groups.push(currentGroup);
        currentGroup = this.createNewGroup(currentLine, groups.length + 1);
      } else {
        this.addLineToGroup(currentGroup, currentLine);
      }
    }

    groups.push(currentGroup);
    return groups;
  }

  /**
   * 创建新的代码组
   * @param {Object} line - 代码行
   * @param {number} id - 组ID
   * @returns {Object} 新的代码组
   */
  static createNewGroup(line, id) {
    return {
      id,
      lines: [line],
      startLine: line.lineNumber,
      endLine: line.lineNumber,
      type: 'added',
    };
  }

  /**
   * 向代码组添加行
   * @param {Object} group - 代码组
   * @param {Object} line - 代码行
   */
  static addLineToGroup(group, line) {
    group.lines.push(line);
    group.endLine = line.lineNumber;
  }

  /**
   * 基础分组判断（只考虑行号连续性）
   * @param {Object} currentLine - 当前行
   * @param {Object} lastLine - 上一行
   * @param {Object} currentGroup - 当前组
   * @returns {boolean} 是否应该开始新组
   */
  static shouldStartNewGroupBasic(currentLine, lastLine, currentGroup) {
    return currentLine.lineNumber !== lastLine.lineNumber + 1;
  }

  /**
   * 判断是否应该开始新组 - 智能语义判断
   * @param {Object} currentLine - 当前行
   * @param {Object} lastLine - 上一行
   * @param {Object} currentGroup - 当前组
   * @returns {boolean} 是否应该开始新组
   */
  static shouldStartNewGroup(currentLine, lastLine, currentGroup) {
    // 1. 行号不连续，必须分组
    if (currentLine.lineNumber !== lastLine.lineNumber + 1) {
      return true;
    }

    // 2. 超过最大行数限制，必须分组
    if (currentGroup.lines.length >= AI_REVIEW_CONFIG.maxLinesPerGroup) {
      return true;
    }

    return false;
  }

  /**
   * 解析 diff 内容，提取变更行信息
   * @param {string} diff - diff 内容
   * @returns {Array} 变更行信息数组
   */
  static parseDiffLines(diff) {
    const lines = diff.split('\n');
    const changeLines = [];
    let currentLineNumber = 0;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];

      if (line.startsWith('@@')) {
        const match = line.match(/@@ -(\d+),?\d* \+(\d+),?\d*/);
        if (match) {
          currentLineNumber = parseInt(match[2]);
        }
        continue;
      }

      if (line.startsWith('+') && !line.startsWith('+++')) {
        const codeContent = line.substring(1);
        changeLines.push({
          type: 'added',
          lineNumber: currentLineNumber,
          content: codeContent,
          originalLine: line,
        });
        currentLineNumber++;
      } else if (line.startsWith('-') && !line.startsWith('---')) {
        continue;
      } else if (!line.startsWith('---') && !line.startsWith('+++')) {
        currentLineNumber++;
      }
    }

    return changeLines;
  }

  /**
   * 工具方法：数组分块
   */
  static chunkArray(array, size) {
    const chunks = [];
    for (let i = 0; i < array.length; i += size) {
      chunks.push(array.slice(i, i + size));
    }
    return chunks;
  }
}

export default CodeProcessor;
