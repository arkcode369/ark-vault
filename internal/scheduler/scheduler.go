package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Job represents a scheduled task.
type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Scheduler runs jobs at fixed intervals.
// It deduplicates hourly jobs so they fire at most once per clock-hour,
// preventing re-execution on bot restarts within the same hour.
type Scheduler struct {
	jobs      []Job
	logger    *slog.Logger
	mu        sync.Mutex
	lastFired map[string]string // job name -> "2006-01-02T15" key
}

// NewScheduler creates a Scheduler.
func NewScheduler(logger *slog.Logger) *Scheduler {
	return &Scheduler{
		logger:    logger,
		lastFired: make(map[string]string),
	}
}

// Add registers a new job.
func (s *Scheduler) Add(job Job) {
	s.jobs = append(s.jobs, job)
}

// Start runs all registered jobs. Blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	for _, job := range s.jobs {
		go s.runJob(ctx, job)
	}
	<-ctx.Done()
}

// hourKey returns a dedup key for the current clock-hour.
func hourKey(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02dT%02d", t.Year(), t.Month(), t.Day(), t.Hour())
}

// MarkAndCheck atomically checks if a job has already fired this hour.
// Returns true if the job should run (first call this hour), false if duplicate.
func (s *Scheduler) MarkAndCheck(jobName string, t time.Time) bool {
	key := hourKey(t)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lastFired[jobName] == key {
		return false
	}
	s.lastFired[jobName] = key
	return true
}

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	s.logger.Info("scheduler: job registered", "name", job.Name, "interval", job.Interval)

	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("scheduler: job stopped", "name", job.Name)
			return
		case <-ticker.C:
			s.logger.Info("scheduler: running job", "name", job.Name)
			if err := job.Run(ctx); err != nil {
				s.logger.Error("scheduler: job failed", "name", job.Name, "error", err)
			}
		}
	}
}
