import dotenv from 'dotenv';
dotenv.config();

const parseBoolean = (val, defaultVal) => {
  if (val === undefined || val === null) return defaultVal;
  return String(val).toLowerCase() === 'true';
};

const parseIntVal = (val, defaultVal) => {
  const parsed = parseInt(val, 10);
  return isNaN(parsed) ? defaultVal : parsed;
};

const config = {
  // 服务器基础配置
  server: {
    port: parseIntVal(process.env.PORT, 3001),
    host: process.env.HOST || '0.0.0.0',
    nodeEnv: process.env.NODE_ENV || 'development',
  },

  // 日志配置
  logger: {
    level: process.env.LOG_LEVEL || 'info',
    enableDebug: parseBoolean(process.env.ENABLE_DEBUG_LOGS, false),
  },

  // GitLab 平台配置
  gitlab: {
    url: process.env.GITLAB_URL || 'https://gitlab.com',
    token: process.env.BOT_TOKEN,
    timeout: parseIntVal(process.env.GITLAB_TIMEOUT, 10000),
    maxRetries: parseIntVal(process.env.GITLAB_MAX_RETRIES, 3),
  },

  // GitHub 平台配置
  github: {
    token: process.env.GITHUB_TOKEN,
    timeout: parseIntVal(process.env.GITHUB_TIMEOUT, 10000),
    maxRetries: parseIntVal(process.env.GITHUB_MAX_RETRIES, 3),
  },

  // AI 服务配置
  ai: {
    provider: process.env.AI_PROVIDER || 'openai', // openai, anthropic
    apiKey: process.env.AI_API_KEY,
    apiUrl: process.env.AI_API_URL || 'https://dashscope.aliyuncs.com/compatible-mode/v1',
    model: process.env.AI_MODEL || 'qwen-plus',
    maxTokens: parseIntVal(process.env.AI_MAX_TOKENS, 2000),
    temperature: parseFloat(process.env.AI_TEMPERATURE) || 0.3,
    requestTimeout: parseIntVal(process.env.AI_TIMEOUT, 30000),
    maxConcurrent: parseIntVal(process.env.MAX_CONCURRENT_AI, 10),
  },

  // 代码审查策略配置
  review: {
    ignoreCrKeyword: 'no-cr',
    maxLinesPerBatch: parseIntVal(process.env.MAX_LINES_PER_BATCH, 50),
    maxFilesConcurrent: parseIntVal(process.env.MAX_FILES_CONCURRENT, 3),
    minLinesToReview: parseIntVal(process.env.MIN_LINES_TO_REVIEW, 3),
    enableInlineComments: parseBoolean(process.env.ENABLE_INLINE_COMMENTS, true),
    addSummaryComment: parseBoolean(process.env.ADD_SUMMARY_COMMENT, true),

    // 性能相关
    maxGroupsPerBatch: parseIntVal(process.env.MAX_GROUPS_PER_BATCH, 15),
    maxLinesPerGroup: parseIntVal(process.env.MAX_LINES_PER_GROUP, 1000),

    // 缓存配置
    cache: {
      expiry: 24 * 60 * 60 * 1000, // 24小时
      maxSize: 1000,
    },

    // 文件过滤规则
    ignoredExtensions: [
      '.css',
      '.scss',
      '.sass',
      '.less',
      '.styl',
      '.md',
      '.markdown',
      '.mdx',
      '.txt',
      '.rst',
      '.adoc',
      '.doc',
      '.docx',
      '.pdf',
      '.toml',
      '.ini',
      '.conf',
      '.cfg',
      '.config',
      '.xml',
      '.json',
      '.map',
    ],

    // 特殊文件处理
    specialHandling: {
      packageFiles: {
        patterns: ['package.json', 'package-lock.json', 'pnpm-lock.yaml', 'yarn.lock'],
        action: 'skip', // 默认跳过包管理文件，除非显式要求
      },
      vueFiles: {
        patterns: ['.vue'],
        action: 'skipStyle',
      },
    },

    // 代码预过滤模式 (Regex)
    skipPatterns: [
      /^\s*\/\/\s*TODO:/i,
      /^\s*\/\/\s*FIXME:/i,
      /^\s*\/\/\s*NOTE:/i,
      /^\s*\/\*.*\*\/\s*$/,
      /^\s*console\.log\s*\(/i,
      /^\s*console\.warn\s*\(/i,
      /^\s*console\.error\s*\(/i,
      /^\s*debugger\s*;?\s*$/,
      /^\s*import\s+.*\s+from\s+['"]/,
      /^\s*export\s+/,
    ],

    // 低质量评论过滤模式
    lowQualityPatterns: [
      'PASS',
      'pass',
      'Pass',
      '无问题',
      '没问题',
      '代码正确',
      '语法正确',
      '命名规范',
      '逻辑合理',
      '没有发现',
      '看起来不错',
      '代码很好',
    ],

    // 冗长评论模式
    verbosePatterns: [
      '新增代码逻辑清晰',
      '结构合理',
      '符合.*规范',
      '不存在明显.*问题',
      '使用合理',
      '整体设计',
      '简洁有效',
      '代码质量良好',
    ],

    // 无意义词汇
    emptyWords: ['逻辑', '结构', '规范', '问题', '合理', '清晰', '良好', '正确'],
  },
};

export default config;
