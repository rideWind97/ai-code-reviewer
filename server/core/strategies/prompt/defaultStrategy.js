import PromptStrategy from '../../interfaces/promptStrategy';

class DefaultPromptStrategy extends PromptStrategy {
  async getInstructions(projectId, ref) {
    let prompt = `请按照以下规则进行专业代码审查:\n\n`;
    prompt += `**一、基本代码质量规则（AI 可自动检测）**\n`;
    prompt += `1. 命名规范：变量、函数、类名应具有描述性，避免缩写或模糊命名\n`;
    prompt += `2. 函数职责单一：一个函数只做一件事，避免超过100行代码\n`;
    prompt += `3. 避免重复代码（DRY）：识别相似代码块，提示提取为公共函数\n`;
    prompt += `4. 遵循项目命名约定（如：camelCase, snake_case, PascalCase）\n`;
    prompt += `5. 注释内容和todo内容需要校验：技术债务和todo需要明确责任人标记，临时方案需要记录原因\n`;
    prompt += `6. 布尔变量建议以 is, has, can 开头\n`;
    prompt += `7. 函数参数建议不超过3个，否则考虑使用对象封装\n`;
    prompt += `8. 严禁使用魔术数\n`;

    prompt += `**二、安全相关规则（AI 可重点扫描）**\n`;
    prompt += `1. 敏感信息：禁止硬编码密码、API Key、密钥等\n\n`;

    prompt += `**示例输出格式：**\n`;
    prompt += `🔍 [AI Review] 建议：\n`;
    prompt += `- 函数 \`processUserData\` 长达80行，建议拆分为多个小函数。\n`;
    prompt += `- 变量名 \`res\` 不够清晰，建议改为 \`userDataResponse\`。\n`;

    prompt += `回复格式：\n1. [具体改进建议或PASS]\n...\n\n`;
    prompt += `要求：中文，每意见<100字，无问题直接回复PASS，不要生成"无问题"、"代码很好"等无意义的评论`;

    return prompt;
  }
}

export default DefaultPromptStrategy;
