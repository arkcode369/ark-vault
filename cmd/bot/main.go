package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arkcode369/ark-vault/internal/adapter/exporter"
	"github.com/arkcode369/ark-vault/internal/adapter/notion"
	"github.com/arkcode369/ark-vault/internal/adapter/telegram"
	badgerdb "github.com/arkcode369/ark-vault/internal/adapter/badger"
	"github.com/arkcode369/ark-vault/internal/config"
	"github.com/arkcode369/ark-vault/internal/domain"
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

	// BadgerDB for gamification
	store, err := badgerdb.OpenStore(cfg.BadgerDBPath)
	if err != nil {
		logger.Error("failed to open badger store", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Services
	journalSvc := service.NewJournalService(tradeRepo, memberRepo, imageRepo)
	leaderboardSvc := service.NewLeaderboardService(tradeRepo, memberRepo)
	reportSvc := service.NewReportService(tradeRepo, memberRepo)
	gamRepo := badgerdb.NewGamificationRepo(store)
	gamSvc := service.NewGamificationService(gamRepo, tradeRepo, memberRepo)

	// Badge & Challenge services
	badgeRepo := badgerdb.NewBadgeRepo(store)
	challengeRepo := badgerdb.NewChallengeRepo(store)
	badgeSvc := service.NewBadgeService(badgeRepo, tradeRepo, gamSvc)
	challengeSvc := service.NewChallengeService(challengeRepo, tradeRepo, memberRepo, gamSvc, badgeSvc)

	// Exporter (CSV + PDF)
	exp := exporter.NewExporter()

	// Rate limiter
	limiter := telegram.NewRateLimiter(cfg.RateLimitPerMin, 1*time.Minute)

	// Telegram
	sender := telegram.NewSender(cfg.TelegramToken)
	svc := telegram.Services{
		Journal:      journalSvc,
		Leaderboard:  leaderboardSvc,
		Report:       reportSvc,
		Gamification: gamSvc,
		Badge:        badgeSvc,
		Challenge:    challengeSvc,
	}
	handler := telegram.NewHandler(
		sender, svc, exp, tradeRepo, memberRepo, limiter, logger,
		cfg.CommunityGroupID, cfg.OwnerID,
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

	// Scheduler
	sched := scheduler.NewScheduler(logger)

	// Weekly report auto-post
	if cfg.ReportChatID != 0 {
		sched.Add(scheduler.Job{
			Name:     "weekly-report",
			Interval: 1 * time.Hour, // check hourly
			Run: func(ctx context.Context) error {
				now := time.Now().UTC()
				dayMatch := isDayOfWeek(now, cfg.ReportDay)
				hourMatch := now.Hour() == cfg.ReportHour
				if dayMatch && hourMatch {
					logger.Info("posting scheduled weekly report", "chat_id", cfg.ReportChatID, "thread_id", cfg.ReportThreadID)
					return handler.SendScheduledReport(ctx, cfg.ReportChatID, cfg.ReportThreadID)
				}
				return nil
			},
		})
		logger.Info("scheduler: weekly-report enabled", "report_day", cfg.ReportDay, "report_hour", cfg.ReportHour)
	}

	// Challenge scheduler jobs
	sched.Add(scheduler.Job{
		Name:     "challenge-finalize",
		Interval: 1 * time.Hour,
		Run: func(ctx context.Context) error {
			now := time.Now().UTC()
			// Finalize on Sunday at report hour
			if now.Weekday() == time.Sunday && now.Hour() == cfg.ReportHour {
				// Get last week's challenge
				lastWeek := now.AddDate(0, 0, -7)
				yearWeek := domain.YearWeekString(lastWeek)
				results, err := challengeSvc.FinalizeChallenge(ctx, yearWeek)
				if err != nil {
					return err
				}
				if len(results) > 0 && cfg.ReportChatID != 0 {
					challenge, _ := challengeSvc.GetOrCreateChallenge(ctx, lastWeek)
					if challenge != nil {
						text := telegram.FormatChallengeResults(challenge, results)
						sender.SendHTML(ctx, cfg.ReportChatID, text, cfg.ReportThreadID)
					}
				}
			}
			return nil
		},
	})

	sched.Add(scheduler.Job{
		Name:     "challenge-announce",
		Interval: 1 * time.Hour,
		Run: func(ctx context.Context) error {
			now := time.Now().UTC()
			// Announce on Monday at 8
			if now.Weekday() == time.Monday && now.Hour() == 8 {
				challenge, err := challengeSvc.GetOrCreateChallenge(ctx, now)
				if err != nil {
					return err
				}
				if cfg.ReportChatID != 0 {
					text := fmt.Sprintf("⚔️ <b>New Weekly Challenge!</b>\n\n<b>%s</b>\n%s\n\nGunakan /challenge untuk melihat standings.", challenge.Title, challenge.Description)
					sender.SendHTML(ctx, cfg.ReportChatID, text, cfg.ReportThreadID)
				}
			}
			return nil
		},
	})

	go sched.Start(ctx)
	logger.Info("scheduler started")

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
