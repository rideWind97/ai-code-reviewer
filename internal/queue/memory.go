package queue

import (
	"context"
	"sync"
)

// Job 表示队列中的一个任务。
//
// Payload 保留原始 webhook body，后续 worker 可以重新解析出 repo、PR 编号等信息。
type Job struct {
	ID      string // 幂等 ID；同一个 ID 重复入队时会被忽略。
	Type    string // 任务类型；当前只有 "review"，后续可扩展。
	Payload []byte // 任务原始数据。
}

// Queue 是队列抽象。
//
// MVP 使用内存实现；生产环境可以实现同样接口，替换为 PostgreSQL、Redis 或 SQS。
type Queue interface {
	Enqueue(context.Context, Job) error
	Dequeue(context.Context) (Job, bool, error)
}

// Memory 是进程内队列实现。
//
// 它适合测试和本地跑通流程，不适合生产：进程重启后任务会丢失。
type Memory struct {
	mu       sync.Mutex          // 保护下面三个字段，避免并发读写 map/slice。
	jobs     []Job               // 保存已入队任务。
	seen     map[string]struct{} // 记录已经入过队的 job ID，用于幂等。
	inFlight map[string]struct{} // 记录已被 worker 取走但尚未完成的 job ID。
}

// NewMemory 创建一个空的内存队列。
func NewMemory() *Memory {
	return &Memory{
		seen:     make(map[string]struct{}),
		inFlight: make(map[string]struct{}),
	}
}

// Enqueue 入队任务。
//
// 如果 job.ID 已经出现过，直接返回 nil，表示“重复事件已被幂等处理”。
func (m *Memory) Enqueue(_ context.Context, job Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.seen[job.ID]; ok {
		return nil
	}
	m.seen[job.ID] = struct{}{}
	m.jobs = append(m.jobs, job)
	return nil
}

// Dequeue 取出一个还没有被取走的任务。
//
// 返回值 ok=false 表示当前没有可处理任务。
func (m *Memory) Dequeue(_ context.Context) (Job, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, job := range m.jobs {
		if _, ok := m.inFlight[job.ID]; ok {
			continue
		}
		m.inFlight[job.ID] = struct{}{}
		return job, true, nil
	}
	return Job{}, false, nil
}

// Jobs 返回当前队列快照，主要用于测试断言。
func (m *Memory) Jobs() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()

	jobs := make([]Job, len(m.jobs))
	copy(jobs, m.jobs)
	return jobs
}
