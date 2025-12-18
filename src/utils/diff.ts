import parseDiff from 'parse-diff';

/**
 * Checks if a file should be skipped based on extension.
 */
export function shouldSkipFile(filename: string): boolean {
  const ignoreExtensions = [
    '.lock',
    '.min.js',
    '.map',
    '.png',
    '.jpg',
    '.jpeg',
    '.svg',
    '.gif',
    '.ico',
    '.pdf',
    '.woff',
    '.woff2',
    '.ttf',
    '.eot',
    '.mp4',
    '.webm',
    '.zip',
    '.tar',
    '.gz',
    '.7z',
    '.rar',
    '.exe',
    '.dll',
    '.so',
    '.dylib',
  ];
  return ignoreExtensions.some((ext) => filename.endsWith(ext));
}

/**
 * Returns a set of line numbers in the new file that are present in the diff.
 * This includes added lines and context lines.
 *
 * Platforms typically only allow commenting on lines that appear in the diff chunk.
 */
export function getReviewableLines(diff: string): Set<number> {
  const validLines = new Set<number>();

  try {
    const files = parseDiff(diff);
    for (const file of files) {
      for (const chunk of file.chunks) {
        for (const change of chunk.changes) {
          // For 'add' and 'normal' (context), we have a line number in the new file (ln2).
          // For 'del', we generally can't comment on the *new* file side for that line,
          // unless we are commenting on the context around it.
          // We stick to ln2 which represents the line in the new version.
          if (change.type === 'normal') {
            validLines.add(change.ln2);
          } else if (change.type === 'add') {
            validLines.add(change.ln);
          }
        }
      }
    }
  } catch (e) {
    console.error('Error parsing diff:', e);
  }

  return validLines;
}
