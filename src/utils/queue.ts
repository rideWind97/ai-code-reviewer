import PQueue from 'p-queue';

// Global queue to limit concurrent AI analysis across all requests
// This helps prevent hitting API rate limits (OpenAI/GitHub) and OOM
export const analysisQueue = new PQueue({
  concurrency: parseInt(process.env.CONCURRENCY_LIMIT || '5'),
});
