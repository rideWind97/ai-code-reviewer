export default class Logger {
  static info(message, meta) {
    console.log('[INFO]', message, meta || '');
  }
  static warn(message, meta) {
    console.warn('[WARN]', message, meta || '');
  }
  static error(message, meta) {
    console.error('[ERROR]', message, meta || '');
  }
  static startTimer(label) {
    return { label, start: Date.now() };
  }
  static endTimer(label, timer, meta) {
    if (!timer) return;
    const duration = Date.now() - timer.start;
    console.log(`[TIMER] ${label}: ${duration}ms`, meta || '');
  }
}
