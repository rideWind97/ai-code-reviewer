package queue

import (
	"context"
	"sync"
)

type Job struct {
	ID      string
	Type    string
	Payload []byte
}

type Queue interface {
	Enqueue(context.Context, Job) error
	Dequeue(context.Context) (Job, bool, error)
}

type Memory struct {
	mu       sync.Mutex
	jobs     []Job
	seen     map[string]struct{}
	inFlight map[string]struct{}
}

func NewMemory() *Memory {
	return &Memory{
		seen:     make(map[string]struct{}),
		inFlight: make(map[string]struct{}),
	}
}

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

func (m *Memory) Jobs() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()

	jobs := make([]Job, len(m.jobs))
	copy(jobs, m.jobs)
	return jobs
}
