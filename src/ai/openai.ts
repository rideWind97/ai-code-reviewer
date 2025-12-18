import OpenAI from 'openai';
import { IAIProvider, FileChange, ReviewContext, AIReviewResult } from '../interfaces';
import { config } from '../config';

export class OpenAIProvider implements IAIProvider {
  private client: OpenAI;

  constructor() {
    this.client = new OpenAI({
      apiKey: config.ai.apiKey,
      baseURL: config.ai.baseUrl, // Optional for compatible APIs
    });
  }

  async analyzeCode(fileChange: FileChange, context: ReviewContext): Promise<AIReviewResult[]> {
    if (!fileChange.rawDiff) return [];

    const defaultPrompt = `
You are an expert code reviewer.
Your task is to review the provided code diff and identify potential bugs, security issues, and bad practices.
Output your review in a strictly JSON format (array of objects).
Do not output "nitpicks" or trivial style suggestions unless they are critical.
The output format should be:
[
  {
    "lineNumber": <number>, // The line number in the new file
    "reviewContent": "<string>", // The review comment
    "suggestion": "<string>", // Optional code suggestion
    "severity": "info" | "warning" | "critical"
  }
]
    `.trim();

    const systemPrompt = `${context.customPrompt || config.ai.systemPrompt || defaultPrompt}
Language: ${context.language}`;

    const userPrompt = `
File: ${fileChange.filename}
Diff:
${fileChange.rawDiff}
    `.trim();

    try {
      const response = await this.client.chat.completions.create({
        model: config.ai.model || 'gpt-4o',
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: userPrompt },
        ],
        response_format: { type: 'json_object' },
      });

      const content = response.choices[0].message.content;
      if (!content) return [];

      const result = JSON.parse(content);
      // Handle case where LLM might wrap it in a key like "reviews"
      const reviews = Array.isArray(result) ? result : result.reviews || [];

      return reviews.map((r: any) => ({
        lineNumber: r.lineNumber,
        reviewContent: r.reviewContent,
        suggestion: r.suggestion,
        severity: r.severity || 'info',
      }));
    } catch (error) {
      console.error('AI Analysis failed:', error);
      return [];
    }
  }
}
