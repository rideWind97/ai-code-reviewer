export interface IPlatformProvider {
  parseEvent(request: any): Promise<ReviewEvent | null>;
  getChanges(event: ReviewEvent): Promise<FileChange[]>;
  postReview(event: ReviewEvent, comments: ReviewComment[]): Promise<void>;
  getReviewComments(event: ReviewEvent): Promise<ReviewComment[]>;
}

export interface IAIProvider {
  analyzeCode(fileChange: FileChange, context: ReviewContext): Promise<AIReviewResult[]>;
}

export interface ReviewEvent {
  platform: 'github' | 'gitlab' | 'gitea';
  eventType: 'pull_request' | 'merge_request';
  eventId: string; // PR Number or IID
  projectId?: string | number; // For GitLab/Gitea (Repo ID)
  repoName: string; // owner/repo
  owner: string;
  commitSha: string; // Head SHA
  baseSha: string; // Base/Target SHA (Required for GitLab position)
  installationId?: number;
  isDraft?: boolean;
}

export interface FileChange {
  filename: string;
  status: 'added' | 'modified' | 'deleted' | 'renamed';
  rawDiff: string;
  patch?: string;
  blobUrl?: string;
}

export interface ReviewComment {
  path: string;
  line: number;
  body: string;
  side?: 'LEFT' | 'RIGHT';
}

export interface ReviewContext {
  language: string;
  customPrompt?: string;
}

export interface AIReviewResult {
  lineNumber: number;
  reviewContent: string;
  suggestion?: string;
  severity: 'info' | 'warning' | 'critical';
}
