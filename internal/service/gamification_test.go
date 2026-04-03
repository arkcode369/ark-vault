package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// mockGamificationStore is a test double for GamificationStore
type mockGamificationStore struct {
	xpLogs   map[int64][]domain.XPLog
	streaks  map[int64]domain.Streak
	lifetime map[int64]domain.LifetimeXP
}

func newMockGamificationStore() *mockGamificationStore {
	return &mockGamificationStore{
		xpLogs:   make(map[int64][]domain.XPLog),
		streaks:  make(map[int64]domain.Streak),
		lifetime: make(map[int64]domain.LifetimeXP),
	}
}

func (m *mockGamificationStore) GetXPLogs(ctx context.Context, memberID int64, since time.Time) ([]domain.XPLog, error) {
	return m.xpLogs[memberID], nil
}

func (m *mockGamificationStore) AppendXPLog(ctx context.Context, memberID int64, log domain.XPLog) error {
	m.xpLogs[memberID] = append(m.xpLogs[memberID], log)
	return nil
}

func (m *mockGamificationStore) GetStreak(ctx context.Context, memberID int64) (domain.Streak, error) {
	if s, ok := m.streaks[memberID]; ok {
		return s, nil
	}
	return domain.Streak{}, nil
}

func (m *mockGamificationStore) SetStreak(ctx context.Context, memberID int64, s domain.Streak) error {
	m.streaks[memberID] = s
	return nil
}

func (m *mockGamificationStore) GetLifetimeXP(ctx context.Context, memberID int64) (domain.LifetimeXP, error) {
	if l, ok := m.lifetime[memberID]; ok {
		return l, nil
	}
	return domain.LifetimeXP{}, nil
}

func (m *mockGamificationStore) SetLifetimeXP(ctx context.Context, memberID int64, l domain.LifetimeXP) error {
	m.lifetime[memberID] = l
	return nil
}

// mockTradeRepo for testing
type mockGamificationTradeRepo struct {
	trades map[int64][]domain.Trade
}

func newMockGamificationTradeRepo() *mockGamificationTradeRepo {
	return &mockGamificationTradeRepo{trades: make(map[int64][]domain.Trade)}
}

func (m *mockGamificationTradeRepo) GetTrades(ctx context.Context, memberID int64) ([]domain.Trade, error) {
	return m.trades[memberID], nil
}

func (m *mockGamificationTradeRepo) SaveTrade(ctx context.Context, memberID int64, trade domain.Trade) error {
	m.trades[memberID] = append(m.trades[memberID], trade)
	return nil
}

// mockMemberRepo for testing
type mockGamificationMemberRepo struct {
	members map[int64]*domain.Member
}

func newMockGamificationMemberRepo() *mockGamificationMemberRepo {
	return &mockGamificationMemberRepo{members: make(map[int64]*domain.Member)}
}

func (m *mockGamificationMemberRepo) GetMember(ctx context.Context, tgID int64) (*domain.Member, error) {
	if member, ok := m.members[tgID]; ok {
		return member, nil
	}
	return nil, errors.New("member not found")
}

func (m *mockGamificationMemberRepo) SaveMember(ctx context.Context, member *domain.Member) error {
	m.members[member.TelegramID] = member
	return nil
}

func (m *mockGamificationMemberRepo) GetOrCreateMember(ctx context.Context, tgID int64, username string) (*domain.Member, error) {
	if member, ok := m.members[tgID]; ok {
		return member, nil
	}
	member := &domain.Member{TelegramID: tgID, Username: username}
	m.members[tgID] = member
	return member, nil
}

func TestGamificationService_OnTradeResult_Win(t *testing.T) {
	store := newMockGamificationStore()
	tradesRepo := newMockGamificationTradeRepo()
	membersRepo := newMockGamificationMemberRepo()

	svc := NewGamificationService(store, tradesRepo, membersRepo)

	ctx := context.Background()
	memberID := int64(12345)

	trade := domain.Trade{
		ID:       "trade-1",
		Date:     time.Now(),
		Symbol:   "BTC",
		Status:   domain.Win,
		ResultRR: 2.5,
	}

	result, err := svc.OnTradeResult(ctx, memberID, trade)
	if err != nil {
		t.Fatalf("OnTradeResult failed: %v", err)
	}

	// Verify XP gained for a win
	if result.XPGained <= 0 {
		t.Errorf("Expected positive XP for win, got %d", result.XPGained)
	}

	// Verify level calculation
	if result.Level < 1 {
		t.Errorf("Expected level >= 1, got %d", result.Level)
	}

	// Verify streak increased
	if result.Streak < 1 {
		t.Errorf("Expected streak >= 1, got %d", result.Streak)
	}
}

func TestGamificationService_OnTradeResult_Loss(t *testing.T) {
	store := newMockGamificationStore()
	tradesRepo := newMockGamificationTradeRepo()
	membersRepo := newMockGamificationMemberRepo()

	svc := NewGamificationService(store, tradesRepo, membersRepo)

	ctx := context.Background()
	memberID := int64(12345)

	trade := domain.Trade{
		ID:       "trade-1",
		Date:     time.Now(),
		Symbol:   "BTC",
		Status:   domain.Loss,
		ResultRR: -1.0,
	}

	result, err := svc.OnTradeResult(ctx, memberID, trade)
	if err != nil {
		t.Fatalf("OnTradeResult failed: %v", err)
	}

	// Verify XP gained for a loss (should be less than win)
	if result.XPGained < 0 {
		t.Errorf("Expected non-negative XP for loss, got %d", result.XPGained)
	}
}

func TestGamificationService_StreakContinuation(t *testing.T) {
	store := newMockGamificationStore()
	tradesRepo := newMockGamificationTradeRepo()
	membersRepo := newMockGamificationMemberRepo()

	svc := NewGamificationService(store, tradesRepo, membersRepo)

	ctx := context.Background()
	memberID := int64(12345)

	// Set initial streak
	store.streaks[memberID] = domain.Streak{
		Current: 3,
		Best:    5,
	}

	trade := domain.Trade{
		ID:       "trade-1",
		Date:     time.Now(),
		Symbol:   "BTC",
		Status:   domain.Win,
		ResultRR: 1.5,
	}

	result, err := svc.OnTradeResult(ctx, memberID, trade)
	if err != nil {
		t.Fatalf("OnTradeResult failed: %v", err)
	}

	// Streak should continue from previous value
	if result.Streak <= 3 {
		t.Errorf("Expected streak to continue from 3, got %d", result.Streak)
	}
}

func TestCalculateLevelAndTitle(t *testing.T) {
	tests := []struct {
		xp           int
		wantLevel    int
		wantTitleMin string // minimum title to expect
	}{
		{0, 1, "Novice"},
		{100, 1, "Novice"},
		{500, 2, "Apprentice"},
		{1500, 4, "Journeyman"},
		{5000, 7, "Analyst"},
		{20000, 11, "Fund Manager"},
	}

	for _, tt := range tests {
		level, title := calculateLevelAndTitle(tt.xp)
		if level != tt.wantLevel {
			t.Errorf("calculateLevelAndTitle(%d) level = %d, want %d", tt.xp, level, tt.wantLevel)
		}
		if title == "" {
			t.Errorf("calculateLevelAndTitle(%d) title should not be empty", tt.xp)
		}
	}
}

func TestGamificationService_GetStats(t *testing.T) {
	store := newMockGamificationStore()
	tradesRepo := newMockGamificationTradeRepo()
	membersRepo := newMockGamificationMemberRepo()

	svc := NewGamificationService(store, tradesRepo, membersRepo)

	ctx := context.Background()
	memberID := int64(12345)

	// Pre-populate some XP
	store.lifetime[memberID] = domain.LifetimeXP{Total: 1000}
	store.streaks[memberID] = domain.Streak{Current: 5, Best: 10}

	stats, err := svc.GetStats(ctx, memberID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalXP != 1000 {
		t.Errorf("Expected TotalXP = 1000, got %d", stats.TotalXP)
	}

	if stats.CurrentStreak != 5 {
		t.Errorf("Expected CurrentStreak = 5, got %d", stats.CurrentStreak)
	}

	if stats.BestStreak != 10 {
		t.Errorf("Expected BestStreak = 10, got %d", stats.BestStreak)
	}
}
