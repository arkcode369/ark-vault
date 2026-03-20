package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Job represents a scheduled task.
type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Scheduler runs jobs at fixed intervals.
type Scheduler struct {
	jobs   []Job
	logger *slog.Logger
}

// NewScheduler creates a Scheduler.
func NewScheduler(logger *slog.Logger) *Scheduler {
	return &Scheduler{logger: logger}
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

func (s *Scheduler) runJob(ctx context.Context, job Job) {
	s.logger.Info("scheduler: job registered", "name", job.Name, "interval", job.Interval)

	// Calculate initial delay to align with the next interval boundary
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
