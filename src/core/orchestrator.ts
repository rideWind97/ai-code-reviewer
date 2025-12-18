import { IPlatformProvider, IAIProvider, ReviewEvent, ReviewContext } from '../interfaces';
import { getReviewableLines, shouldSkipFile } from '../utils/diff';
import { logger } from '../utils/logger';

export class ReviewOrchestrator {
  constructor(
    private platform: IPlatformProvider,
    private ai: IAIProvider,
  ) {}

  async handleEvent(event: ReviewEvent) {
    logger.info({ eventId: event.eventId, repo: event.repoName }, 'Processing event');

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

    logger.info(
      { filesChanged: files.length, existingComments: existingComments.length },
      'Changes and comments fetched',
    );

    // 2. Filter & Analyze
    const allComments = [];

    for (const file of files) {
      if (shouldSkipFile(file.filename) || file.status === 'deleted' || !file.rawDiff) {
        continue;
      }

      logger.info({ filename: file.filename }, 'Analyzing file');

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
          logger.debug(
            { line: res.lineNumber },
            'Skipping comment on line as it is not in the diff context',
          );
          continue;
        }

        // Deduplication: Skip if already commented
        if (commentedLines.has(`${file.filename}:${res.lineNumber}`)) {
          logger.info({ file: file.filename, line: res.lineNumber }, 'Skipping duplicate comment');
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
      logger.info({ count: allComments.length }, 'Posting comments');
      await this.platform.postReview(event, allComments);
      logger.info('Comments posted successfully');
    } else {
      logger.info('No issues found or all duplicates filtered.');
    }
  }
}
