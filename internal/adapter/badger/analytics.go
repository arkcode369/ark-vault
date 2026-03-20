package badger

import (
	"context"
	"fmt"
	"time"

	badgerdb "github.com/dgraph-io/badger/v4"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// AnalyticsCacheRepo implements ports.AnalyticsCacheStore backed by BadgerDB.
type AnalyticsCacheRepo struct {
	store *Store
}

// NewAnalyticsCacheRepo creates a new AnalyticsCacheRepo.
func NewAnalyticsCacheRepo(store *Store) *AnalyticsCacheRepo {
	return &AnalyticsCacheRepo{store: store}
}

func analyticsCacheKey(telegramID int64) string {
	return fmt.Sprintf("gam:analytics_cache:%d", telegramID)
}

// GetAnalyticsCache returns cached AI analytics for the given member.
// Returns nil (not an error) when no cache entry exists or it has expired.
func (r *AnalyticsCacheRepo) GetAnalyticsCache(_ context.Context, telegramID int64) (*domain.TradeAnalytics, error) {
	var a domain.TradeAnalytics
	err := r.store.Get(analyticsCacheKey(telegramID), &a)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Double-check expiry in case TTL hasn't purged it yet
	if time.Now().After(a.ExpiresAt) {
		return nil, nil
	}
	return &a, nil
}

// SaveAnalyticsCache persists AI analytics with a 24-hour TTL.
func (r *AnalyticsCacheRepo) SaveAnalyticsCache(_ context.Context, analytics *domain.TradeAnalytics) error {
	return r.store.SetWithTTL(analyticsCacheKey(analytics.TelegramID), analytics, 24*time.Hour)
}

// ReportCardRepo implements ports.ReportCardStore backed by BadgerDB.
type ReportCardRepo struct {
	store *Store
}

// NewReportCardRepo creates a new ReportCardRepo.
func NewReportCardRepo(store *Store) *ReportCardRepo {
	return &ReportCardRepo{store: store}
}

func reportCardKey(telegramID int64, yearMonth string) string {
	return fmt.Sprintf("gam:report:%d:%s", telegramID, yearMonth)
}

// GetReportCard returns the monthly report card for the given member and month.
// Returns nil (not an error) when no report card exists.
func (r *ReportCardRepo) GetReportCard(_ context.Context, telegramID int64, yearMonth string) (*domain.MonthlyReportCard, error) {
	var rc domain.MonthlyReportCard
	err := r.store.Get(reportCardKey(telegramID, yearMonth), &rc)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

// SaveReportCard persists a monthly report card.
func (r *ReportCardRepo) SaveReportCard(_ context.Context, report *domain.MonthlyReportCard) error {
	return r.store.Set(reportCardKey(report.TelegramID, report.YearMonth), report)
}
