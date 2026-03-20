package domain

import "time"

// BadgeID is a unique identifier for a badge type.
type BadgeID string

const (
	BadgeFirstTrade   BadgeID = "first_trade"
	Badge10Trades     BadgeID = "10_trades"
	Badge50Trades     BadgeID = "50_trades"
	Badge100Trades    BadgeID = "100_trades"
	BadgeWinStreak3   BadgeID = "win_streak_3"
	BadgeWinStreak5   BadgeID = "win_streak_5"
	BadgeStreak7      BadgeID = "streak_7"
	BadgeStreak30     BadgeID = "streak_30"
	BadgeFirstGreenMo BadgeID = "first_green_month"
	Badge10RMonth     BadgeID = "10r_month"
	BadgeWinrate60    BadgeID = "winrate_60"
	BadgeChallengeWin BadgeID = "challenge_winner"
	BadgeGoalAchiever BadgeID = "goal_achiever"
)

// BadgeDefinition describes a badge.
type BadgeDefinition struct {
	ID          BadgeID
	Name        string
	Description string
	Emoji       string
}

// BadgeRegistry is the ordered list of all badges.
var BadgeRegistry = []BadgeDefinition{
	{BadgeFirstTrade, "Trade Pertama", "Catat 1 trade", "\U0001f3af"},
	{Badge10Trades, "10 Trade", "Catat 10 trade", "\U0001f4c8"},
	{Badge50Trades, "50 Trade", "Catat 50 trade", "\U0001f4ca"},
	{Badge100Trades, "Centurion", "Catat 100 trade", "\U0001f4af"},
	{BadgeWinStreak3, "3 Win Streak", "3 kemenangan beruntun", "\U0001f525"},
	{BadgeWinStreak5, "5 Win Streak", "5 kemenangan beruntun", "\U0001f525\U0001f525"},
	{BadgeStreak7, "Disiplin 7 Hari", "7 hari jurnal berturut-turut", "\U0001f4c5"},
	{BadgeStreak30, "Disiplin 30 Hari", "30 hari jurnal berturut-turut", "\U0001f5d3\ufe0f"},
	{BadgeFirstGreenMo, "Bulan Hijau Pertama", "Bulan pertama dengan RR positif", "\U0001f7e2"},
	{Badge10RMonth, "10R Month", "Bulan dengan 10+ RR", "\U0001f4b0"},
	{BadgeWinrate60, "Sharp Shooter", "60%+ win rate (min 20 trades)", "\U0001f3af"},
	{BadgeChallengeWin, "Juara Mingguan", "Menang weekly challenge", "\U0001f3c6"},
	{BadgeGoalAchiever, "Goal Crusher", "Selesaikan monthly goal", "\u2705"},
}

// BadgeAward records when a user earned a badge.
type BadgeAward struct {
	TelegramID int64     `json:"telegram_id"`
	BadgeID    BadgeID   `json:"badge_id"`
	AwardedAt  time.Time `json:"awarded_at"`
}

// GetBadgeDefinition returns the definition for a badge ID, or nil.
func GetBadgeDefinition(id BadgeID) *BadgeDefinition {
	for _, b := range BadgeRegistry {
		if b.ID == id {
			return &b
		}
	}
	return nil
}
