package queue

import (
	"context"
	"database/sql"
	"errors"
)

// SQLDB 把 database/sql.DB 包装成 PostgresDB。
//
// PostgresQueue 只依赖很小的 PostgresDB 接口；这个包装器负责真实数据库事务。
type SQLDB struct {
	DB *sql.DB // 标准库数据库连接池。
}

// ExecContext 直接转发到 database/sql。
func (db *SQLDB) ExecContext(ctx context.Context, query string, args ...any) error {
	_, err := db.DB.ExecContext(ctx, query, args...)
	return err
}

// DequeueJob 用事务领取一个 queued job。
//
// SELECT ... FOR UPDATE SKIP LOCKED 可以让多个 worker 并发领取任务时互不阻塞，
// 同一个任务只会被一个 worker 锁住并标记为 running。
func (db *SQLDB) DequeueJob(ctx context.Context) (Job, bool, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return Job{}, false, err
	}
	defer tx.Rollback()

	var job Job
	err = tx.QueryRowContext(ctx, `
SELECT id, job_type, payload
FROM review_jobs
WHERE state = 'queued'
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT 1
`).Scan(&job.ID, &job.Type, &job.Payload)
	if errors.Is(err, sql.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE review_jobs
SET state = 'running', attempts = attempts + 1, locked_at = now(), updated_at = now()
WHERE id = $1
`, job.ID); err != nil {
		return Job{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Job{}, false, err
	}
	return job, true, nil
}
