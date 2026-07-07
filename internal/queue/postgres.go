package queue

import "context"

// PostgresDB 是 PostgresQueue 需要的最小数据库接口。
//
// 真实环境可以用 database/sql 封装实现；测试里用 fake 即可。
type PostgresDB interface {
	ExecContext(ctx context.Context, query string, args ...any) error
	DequeueJob(ctx context.Context) (Job, bool, error)
}

// PostgresQueue 是持久化队列实现。
type PostgresQueue struct {
	db PostgresDB
}

// NewPostgres 创建 Postgres 持久化队列。
func NewPostgres(db PostgresDB) *PostgresQueue {
	return &PostgresQueue{db: db}
}

// Enqueue 使用 ON CONFLICT DO NOTHING 保证幂等入队。
func (q *PostgresQueue) Enqueue(ctx context.Context, job Job) error {
	return q.db.ExecContext(ctx, `
INSERT INTO review_jobs (id, job_type, payload, state)
VALUES ($1, $2, $3, 'queued')
ON CONFLICT (id) DO NOTHING
`, job.ID, job.Type, job.Payload)
}

// Dequeue 从数据库中领取一个 queued job。
func (q *PostgresQueue) Dequeue(ctx context.Context) (Job, bool, error) {
	return q.db.DequeueJob(ctx)
}
