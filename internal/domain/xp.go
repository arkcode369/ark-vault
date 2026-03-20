package domain

// XP amounts for various actions
const (
	XPTradeLogged       = 10
	XPWinBonus          = 5
	XPWinHighRR         = 10 // replaces XPWinBonus for 2R+
	XPDailyStreak       = 5
	XPStreak7           = 25
	XPStreak30          = 100
	XPBadgeEarned       = 20
	XPChallengeParticip = 10
	XPChallenge1st      = 50
	XPChallenge2nd      = 30
	XPChallenge3rd      = 20
	XPGoalAchieved      = 75
	XPScreenshot        = 3
)

// LevelThreshold defines a level and its XP requirement.
type LevelThreshold struct {
	Level int
	XP    int
	Title string
}

// LevelTable is the ordered list of level thresholds.
var LevelTable = []LevelThreshold{
	{1, 0, "Pemula"},
	{2, 50, "Pengamat"},
	{3, 150, "Pelajar"},
	{4, 350, "Praktisi"},
	{5, 600, "Trader Aktif"},
	{6, 1000, "Trader Disiplin"},
	{7, 1500, "Trader Konsisten"},
	{8, 2200, "Veteran"},
	{9, 3000, "Master Trader"},
	{10, 4000, "Legenda"},
}

// LevelForXP returns the level and title for the given XP amount.
func LevelForXP(xp int) (int, string) {
	level := 1
	title := "Pemula"
	for _, lt := range LevelTable {
		if xp >= lt.XP {
			level = lt.Level
			title = lt.Title
		} else {
			break
		}
	}
	return level, title
}

// XPForNextLevel returns the XP needed for the next level, or 0 if max level.
func XPForNextLevel(currentXP int) int {
	for _, lt := range LevelTable {
		if lt.XP > currentXP {
			return lt.XP
		}
	}
	return 0
}
