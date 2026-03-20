package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkcode369/ark-vault/internal/adapter/exporter"
	"github.com/arkcode369/ark-vault/internal/adapter/notion"
	"github.com/arkcode369/ark-vault/internal/adapter/telegram"
	"github.com/arkcode369/ark-vault/internal/config"
	"github.com/arkcode369/ark-vault/internal/scheduler"
	"github.com/arkcode369/ark-vault/internal/service"
)

func main() {
	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Config
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Notion adapters
	notionClient := notion.NewClient(cfg.NotionToken)
	memberRepo := notion.NewMemberRepo(notionClient, cfg.NotionParentID)
	tradeRepo := notion.NewTradeRepo(notionClient, memberRepo)
	imageRepo := notion.NewImageRepo(notionClient)

	// Services
	journalSvc := service.NewJournalService(tradeRepo, memberRepo, imageRepo)
	leaderboardSvc := service.NewLeaderboardService(tradeRepo, memberRepo)
	reportSvc := service.NewReportService(tradeRepo, memberRepo)

	// Exporter (CSV + PDF)
	exp := exporter.NewExporter()

	// Rate limiter
	limiter := telegram.NewRateLimiter(cfg.RateLimitPerMin, 1*time.Minute)

	// Telegram
	sender := telegram.NewSender(cfg.TelegramToken)
	handler := telegram.NewHandler(
		sender, journalSvc, leaderboardSvc, reportSvc,
		exp, tradeRepo, memberRepo, limiter, logger,
	)
	bot := telegram.NewBot(cfg.TelegramToken, handler, logger)
	handler.SetBot(bot)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutdown signal received")
		cancel()
	}()

	// Scheduler: weekly report auto-post
	if cfg.ReportChatID != 0 {
		sched := scheduler.NewScheduler(logger)
		sched.Add(scheduler.Job{
			Name:     "weekly-report",
			Interval: 1 * time.Hour, // check hourly
			Run: func(ctx context.Context) error {
				now := time.Now().UTC()
				dayMatch := isDayOfWeek(now, cfg.ReportDay)
				hourMatch := now.Hour() == cfg.ReportHour
				if dayMatch && hourMatch {
					logger.Info("posting scheduled weekly report", "chat_id", cfg.ReportChatID)
					return handler.SendScheduledReport(ctx, cfg.ReportChatID)
				}
				return nil
			},
		})
		go sched.Start(ctx)
		logger.Info("scheduler started", "report_day", cfg.ReportDay, "report_hour", cfg.ReportHour)
	}

	logger.Info("ark-vault bot starting")
	if err := bot.Start(ctx); err != nil && err != context.Canceled {
		logger.Error("bot stopped with error", "error", err)
		os.Exit(1)
	}
	logger.Info("bot stopped gracefully")
}

func isDayOfWeek(t time.Time, day string) bool {
	days := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}
	wd, ok := days[day]
	if !ok {
		return false
	}
	return t.Weekday() == wd
}
