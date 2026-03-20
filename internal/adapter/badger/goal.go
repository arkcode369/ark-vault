package badger

import (
	"context"
	"encoding/json"
	"fmt"

	badgerdb "github.com/dgraph-io/badger/v4"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// GoalRepo implements ports.GoalStore backed by BadgerDB.
type GoalRepo struct {
	store *Store
}

// NewGoalRepo creates a new GoalRepo.
func NewGoalRepo(store *Store) *GoalRepo {
	return &GoalRepo{store: store}
}

func goalKey(telegramID int64, yearMonth string) string {
	return fmt.Sprintf("gam:goal:%d:%s", telegramID, yearMonth)
}

const goalPrefix = "gam:goal:"

// GetGoal returns the monthly goal for the given user and month.
// Returns nil (not an error) when no goal exists.
func (r *GoalRepo) GetGoal(_ context.Context, telegramID int64, yearMonth string) (*domain.MonthlyGoal, error) {
	var g domain.MonthlyGoal
	err := r.store.Get(goalKey(telegramID, yearMonth), &g)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// SaveGoal persists a monthly goal.
func (r *GoalRepo) SaveGoal(_ context.Context, goal *domain.MonthlyGoal) error {
	return r.store.Set(goalKey(goal.TelegramID, goal.YearMonth), goal)
}

// ListActiveGoals returns all non-achieved goals for the given yearMonth.
// It scans all goals and filters by yearMonth and achieved status.
func (r *GoalRepo) ListActiveGoals(_ context.Context, yearMonth string) ([]domain.MonthlyGoal, error) {
	raw, err := r.store.Scan(goalPrefix)
	if err != nil {
		return nil, err
	}
	var goals []domain.MonthlyGoal
	for _, b := range raw {
		var g domain.MonthlyGoal
		if err := json.Unmarshal(b, &g); err != nil {
			return nil, err
		}
		if g.YearMonth == yearMonth && !g.Achieved {
			goals = append(goals, g)
		}
	}
	return goals, nil
}
