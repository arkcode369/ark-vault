package service

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// mockChallengeStore is a test double for ChallengeStore
type mockChallengeStore struct {
	challenges map[string]*domain.WeeklyChallenge
}

func newMockChallengeStore() *mockChallengeStore {
	return &mockChallengeStore{challenges: make(map[string]*domain.WeeklyChallenge)}
}

func (m *mockChallengeStore) GetChallenge(ctx context.Context, yearWeek string) (*domain.WeeklyChallenge, error) {
	if c, ok := m.challenges[yearWeek]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockChallengeStore) SaveChallenge(ctx context.Context, c *domain.WeeklyChallenge) error {
	m.challenges[c.YearWeek] = c
	return nil
}

// mockTradeRepo for challenge testing
type mockChallengeTradeRepo struct {
	trades map[int64][]domain.Trade
}

func newMockChallengeTradeRepo() *mockChallengeTradeRepo {
	return &mockChallengeTradeRepo{trades: make(map[int64][]domain.Trade)}
}

func (m *mockChallengeTradeRepo) GetTrades(ctx context.Context, memberID int64) ([]domain.Trade, error) {
	return m.trades[memberID], nil
}

func (m *mockChallengeTradeRepo) SaveTrade(ctx context.Context, memberID int64, trade domain.Trade) error {
	m.trades[memberID] = append(m.trades[memberID], trade)
	return nil
}

// mockMemberRepo for challenge testing
type mockChallengeMemberRepo struct {
	members map[int64]*domain.Member
}

func newMockChallengeMemberRepo() *mockChallengeMemberRepo {
	return &mockChallengeMemberRepo{members: make(map[int64]*domain.Member)}
}

func (m *mockChallengeMemberRepo) GetMember(ctx context.Context, tgID int64) (*domain.Member, error) {
	if member, ok := m.members[tgID]; ok {
		return member, nil
	}
	return nil, nil
}

func (m *mockChallengeMemberRepo) SaveMember(ctx context.Context, member *domain.Member) error {
	m.members[member.TelegramID] = member
	return nil
}

func (m *mockChallengeMemberRepo) GetOrCreateMember(ctx context.Context, tgID int64, username string) (*domain.Member, error) {
	if member, ok := m.members[tgID]; ok {
		return member, nil
	}
	member := &domain.Member{TelegramID: tgID, Username: username}
	m.members[tgID] = member
	return member, nil
}

// Mock badge service for challenge testing
type mockChallengeBadgeService struct{}

func (m *mockChallengeBadgeService) GetMemberStats(ctx context.Context, memberID int64) (*domain.BadgeStats, error) {
	return &domain.BadgeStats{}, nil
}

func TestChallengeService_GetOrCreateChallenge_NewChallenge(t *testing.T) {
	store := newMockChallengeStore()
	tradesRepo := newMockChallengeTradeRepo()
	membersRepo := newMockChallengeMemberRepo()

	// Create minimal gamification and badge services
	gamStore := newMockGamificationStore()
	gamSvc := NewGamificationService(gamStore, tradesRepo, membersRepo)
	badgeSvc := NewBadgeService(nil, tradesRepo, membersRepo)

	svc := NewChallengeService(store, tradesRepo, membersRepo, gamSvc, badgeSvc)

	ctx := context.Background()
	now := time.Now()

	challenge, err := svc.GetOrCreateChallenge(ctx, now)
	if err != nil {
		t.Fatalf("GetOrCreateChallenge failed: %v", err)
	}

	if challenge == nil {
		t.Fatal("Expected challenge to be created, got nil")
	}

	// Verify yearWeek is set correctly
	expectedYearWeek := domain.YearWeekString(now)
	if challenge.YearWeek != expectedYearWeek {
		t.Errorf("Expected YearWeek = %s, got %s", expectedYearWeek, challenge.YearWeek)
	}

	// Verify challenge is saved
	if len(store.challenges) != 1 {
		t.Errorf("Expected 1 challenge in store, got %d", len(store.challenges))
	}
}

func TestChallengeService_GetOrCreateChallenge_ExistingChallenge(t *testing.T) {
	store := newMockChallengeStore()
	tradesRepo := newMockChallengeTradeRepo()
	membersRepo := newMockChallengeMemberRepo()

	gamStore := newMockGamificationStore()
	gamSvc := NewGamificationService(gamStore, tradesRepo, membersRepo)
	badgeSvc := NewBadgeService(nil, tradesRepo, membersRepo)

	svc := NewChallengeService(store, tradesRepo, membersRepo, gamSvc, badgeSvc)

	ctx := context.Background()
	now := time.Now()

	// Pre-populate a challenge
	existingChallenge := &domain.WeeklyChallenge{
		YearWeek: domain.YearWeekString(now),
		Status:   domain.ChallengeActive,
	}
	store.challenges[existingChallenge.YearWeek] = existingChallenge

	challenge, err := svc.GetOrCreateChallenge(ctx, now)
	if err != nil {
		t.Fatalf("GetOrCreateChallenge failed: %v", err)
	}

	if challenge == nil {
		t.Fatal("Expected challenge to be returned, got nil")
	}

	// Should return existing, not create new
	if challenge.Status != domain.ChallengeActive {
		t.Errorf("Expected status = Active, got %v", challenge.Status)
	}
}

func TestChallengeService_GetCurrentWeek(t *testing.T) {
	store := newMockChallengeStore()
	tradesRepo := newMockChallengeTradeRepo()
	membersRepo := newMockChallengeMemberRepo()

	gamStore := newMockGamificationStore()
	gamSvc := NewGamificationService(gamStore, tradesRepo, membersRepo)
	badgeSvc := NewBadgeService(nil, tradesRepo, membersRepo)

	svc := NewChallengeService(store, tradesRepo, membersRepo, gamSvc, badgeSvc)

	ctx := context.Background()

	challenge, err := svc.GetCurrentWeek(ctx)
	if err != nil {
		t.Fatalf("GetCurrentWeek failed: %v", err)
	}

	if challenge == nil {
		t.Fatal("Expected challenge to be created, got nil")
	}

	// Should be for current week
	expectedYearWeek := domain.YearWeekString(time.Now())
	if challenge.YearWeek != expectedYearWeek {
		t.Errorf("Expected YearWeek = %s, got %s", expectedYearWeek, challenge.YearWeek)
	}
}

func TestChallengeService_GetChallengeStats(t *testing.T) {
	store := newMockChallengeStore()
	tradesRepo := newMockChallengeTradeRepo()
	membersRepo := newMockChallengeMemberRepo()

	gamStore := newMockGamificationStore()
	gamSvc := NewGamificationService(gamStore, tradesRepo, membersRepo)
	badgeSvc := NewBadgeService(nil, tradesRepo, membersRepo)

	svc := NewChallengeService(store, tradesRepo, membersRepo, gamSvc, badgeSvc)

	ctx := context.Background()
	now := time.Now()

	// Create a challenge with some participants
	challenge := &domain.WeeklyChallenge{
		YearWeek:     domain.YearWeekString(now),
		Status:       domain.ChallengeActive,
		Participants: []int64{1, 2, 3},
	}
	store.challenges[challenge.YearWeek] = challenge

	stats, err := svc.GetChallengeStats(ctx, now)
	if err != nil {
		t.Fatalf("GetChallengeStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	if stats.ParticipantCount != 3 {
		t.Errorf("Expected 3 participants, got %d", stats.ParticipantCount)
	}
}

func TestChallengeBestRR_NegativeOnlyPortfolio(t *testing.T) {
	// Test ChallengeBestRR with negative-only trades
	trades := []domain.Trade{
		{Status: domain.Loss, ResultRR: -1.0},
		{Status: domain.Loss, ResultRR: -2.5},
		{Status: domain.Loss, ResultRR: -0.5},
	}

	best := ChallengeBestRR(trades)

	// With negative-only portfolio, best should be the least negative (closest to 0)
	expected := -0.5
	if best != expected {
		t.Errorf("ChallengeBestRR with negative trades = %.2f, want %.2f", best, expected)
	}
}

func TestChallengeBestRR_MixedPortfolio(t *testing.T) {
	trades := []domain.Trade{
		{Status: domain.Loss, ResultRR: -1.0},
		{Status: domain.Win, ResultRR: 2.5},
		{Status: domain.Loss, ResultRR: -0.5},
		{Status: domain.Win, ResultRR: 1.5},
	}

	best := ChallengeBestRR(trades)

	expected := 2.5
	if best != expected {
		t.Errorf("ChallengeBestRR with mixed trades = %.2f, want %.2f", best, expected)
	}
}

func TestChallengeBestRR_EmptyTrades(t *testing.T) {
	trades := []domain.Trade{}

	best := ChallengeBestRR(trades)

	// Empty trades should return negative infinity
	if best != math.Inf(-1) {
		t.Errorf("ChallengeBestRR with empty trades = %.2f, want -Inf", best)
	}
}

func TestChallengeBestRR_NoClosedTrades(t *testing.T) {
	trades := []domain.Trade{
		{Status: domain.Open, ResultRR: 0},
		{Status: domain.Pending, ResultRR: 0},
	}

	best := ChallengeBestRR(trades)

	// No closed trades should return negative infinity
	if best != math.Inf(-1) {
		t.Errorf("ChallengeBestRR with no closed trades = %.2f, want -Inf", best)
	}
}
