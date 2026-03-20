package domain

import "time"

// GoalType defines what kind of monthly goal.
type GoalType string

const (
	GoalTotalTrades GoalType = "total_trades"
	GoalTotalRR     GoalType = "total_rr"
	GoalWinRate     GoalType = "win_rate"
	GoalStreakDays  GoalType = "streak_days"
)

// MonthlyGoal stores a member's monthly goal.
type MonthlyGoal struct {
	TelegramID  int64     `json:"telegram_id"`
	YearMonth   string    `json:"year_month"` // "2026-03"
	GoalType    GoalType  `json:"goal_type"`
	TargetValue float64   `json:"target_value"`
	Achieved    bool      `json:"achieved"`
	AchievedAt  time.Time `json:"achieved_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// GoalProgress tracks progress toward a monthly goal.
type GoalProgress struct {
	Goal         *MonthlyGoal
	CurrentValue float64
	Percentage   float64 // 0-100+
}
