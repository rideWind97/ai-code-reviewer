import { IPlatformProvider, IAIProvider, ReviewEvent, ReviewContext } from '../interfaces';
import { getReviewableLines, shouldSkipFile } from '../utils/diff';

export class ReviewOrchestrator {
  constructor(
    private platform: IPlatformProvider,
    private ai: IAIProvider,
  ) {}

  async handleEvent(event: ReviewEvent) {
    console.log(`Processing event ${event.eventId} for ${event.repoName}`);

    // 1. Get Changes & Existing Comments
    const [files, existingComments] = await Promise.all([
      this.platform.getChanges(event),
      this.platform.getReviewComments(event),
    ]);

    // Create a Set of existing comments for fast lookup "path:line"
    const commentedLines = new Set<string>();
    for (const comment of existingComments) {
      commentedLines.add(`${comment.path}:${comment.line}`);
    }

    console.log(
      `Found ${files.length} files changed and ${existingComments.length} existing comments.`,
    );

    // 2. Filter & Analyze
    const allComments = [];

    for (const file of files) {
      if (shouldSkipFile(file.filename) || file.status === 'deleted' || !file.rawDiff) {
        continue;
      }

      console.log(`Analyzing ${file.filename}...`);

      // Parse diff to know which lines are valid for comments
      const validLines = getReviewableLines(file.rawDiff);

      // Define Context
      const context: ReviewContext = {
        language: 'zh-CN',
      };

      // Call AI
      const aiResults = await this.ai.analyzeCode(file, context);

      // Transform AI results to Platform Comments
      for (const res of aiResults) {
        if (res.severity === 'info') continue;

        // Validation: Ensure the line is in the diff
        if (!validLines.has(res.lineNumber)) {
          console.log(
            `Skipping comment on line ${res.lineNumber} as it is not in the diff context.`,
          );
          continue;
        }

        // Deduplication: Skip if already commented
        if (commentedLines.has(`${file.filename}:${res.lineNumber}`)) {
          console.log(`Skipping duplicate comment on ${file.filename}:${res.lineNumber}`);
          continue;
        }

        allComments.push({
          path: file.filename,
          line: res.lineNumber,
          body: `**[AI Review - ${res.severity}]**\n\n${res.reviewContent}\n\n${res.suggestion ? '```suggestion\n' + res.suggestion + '\n```' : ''}`,
          side: 'RIGHT' as const,
        });
      }
    }

    // 3. Post Reviews
    if (allComments.length > 0) {
      console.log(`Posting ${allComments.length} comments...`);
      await this.platform.postReview(event, allComments);
      console.log('Comments posted successfully.');
    } else {
      console.log('No issues found.');
    }
  }
}
