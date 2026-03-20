package telegram

import (
	"fmt"
	"strings"

	"github.com/arkcode369/ark-vault/internal/domain"
	"github.com/arkcode369/ark-vault/internal/service"
)

// FormatStats formats member stats as an HTML Telegram message.
func FormatStats(username string, s *domain.Stats) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 <b>Stats — @%s</b>\n\n", username))
	sb.WriteString(fmt.Sprintf("Total trades: <b>%d</b>\n", s.TotalTrades))
	sb.WriteString(fmt.Sprintf("Win/Loss/BE: <b>%d</b>/%d/%d\n", s.Wins, s.Losses, s.BreakEvens))
	if s.OpenTrades > 0 {
		sb.WriteString(fmt.Sprintf("Open: %d\n", s.OpenTrades))
	}
	sb.WriteString(fmt.Sprintf("Win Rate: <b>%.1f%%</b>\n", s.WinRate))
	sb.WriteString(fmt.Sprintf("Avg RR: <b>%.2f</b>\n", s.AvgRR))
	sb.WriteString(fmt.Sprintf("Total RR: <b>%+.1fR</b>\n", s.TotalRR))
	sb.WriteString(fmt.Sprintf("Best: <b>%+.1fR</b>\n", s.BestRR))
	sb.WriteString(fmt.Sprintf("Worst: <b>%+.1fR</b>\n", s.WorstRR))

	if s.CurStreak > 0 {
		sb.WriteString(fmt.Sprintf("Current streak: 🔥 <b>%d</b> wins\n", s.CurStreak))
	} else if s.CurStreak < 0 {
		sb.WriteString(fmt.Sprintf("Current streak: ❄️ <b>%d</b> losses\n", -s.CurStreak))
	}
	if s.MaxWinStrk > 0 {
		sb.WriteString(fmt.Sprintf("Longest win streak: <b>%d</b>\n", s.MaxWinStrk))
	}

	// Per-asset breakdown
	if len(s.ByAsset) > 1 {
		sb.WriteString("\n<b>Per Asset:</b>\n")
		for at, as := range s.ByAsset {
			sb.WriteString(fmt.Sprintf("  %s: %d trades, %.1f%% WR, %+.1fR\n",
				at.String(), as.Total, as.WinRate, as.TotalRR))
		}
	}

	return sb.String()
}

// FormatLeaderboard formats the leaderboard as an HTML Telegram message.
func FormatLeaderboard(entries []service.LeaderboardEntry, metric string) string {
	var sb strings.Builder
	sb.WriteString("🏆 <b>Leaderboard</b>")
	if metric == "rr" {
		sb.WriteString(" (by RR)")
	} else {
		sb.WriteString(" (by Win Rate)")
	}
	sb.WriteString("\n\n")

	medals := []string{"🥇", "🥈", "🥉"}
	for i, e := range entries {
		rank := fmt.Sprintf("%d.", i+1)
		if i < 3 {
			rank = medals[i]
		}
		name := e.Username
		if name == "" {
			name = e.FirstName
		}
		if metric == "rr" {
			sb.WriteString(fmt.Sprintf("%s <b>@%s</b> — %+.1fR (%d trades)\n",
				rank, name, e.TotalRR, e.TotalTrades))
		} else {
			sb.WriteString(fmt.Sprintf("%s <b>@%s</b> — %.1f%% WR (%d trades)\n",
				rank, name, e.WinRate, e.TotalTrades))
		}
	}

	if len(entries) == 0 {
		sb.WriteString("Belum ada member yang memenuhi syarat.")
	}

	return sb.String()
}

// FormatTradeConfirmation formats a saved trade confirmation message.
func FormatTradeConfirmation(t *domain.Trade) string {
	var sb strings.Builder
	sb.WriteString("✅ <b>Trade berhasil dicatat!</b>\n\n")
	sb.WriteString(fmt.Sprintf("Symbol: <b>%s</b> (%s)\n", t.Symbol, t.AssetType.String()))
	sb.WriteString(fmt.Sprintf("Direction: <b>%s</b>\n", t.Direction))
	sb.WriteString(fmt.Sprintf("Status: <b>%s</b>", t.Status))
	if t.ResultRR != 0 {
		sb.WriteString(fmt.Sprintf(" (%+.1fR)", t.ResultRR))
	}
	sb.WriteString("\n")
	if t.TimeWindow != "" {
		sb.WriteString(fmt.Sprintf("Session: <b>%s</b>\n", t.TimeWindow))
	}
	if t.Confluence != "" {
		sb.WriteString(fmt.Sprintf("Confluence: %s\n", t.Confluence))
	}
	return sb.String()
}

// FormatHelp returns the /help message.
func FormatHelp() string {
	return `🔐 <b>ARK Vault — Help</b>

<b>Mencatat Trade:</b>
• /journal — Guided flow (step-by-step dengan tombol)
• Kirim pesan dengan format:
<pre>#journal
Pair: XAUUSD
Type: BUY
RR: +2
Session: London
Confluence: FVG + OB mitigation on 15m</pre>
  Bisa attach screenshot bersamaan.

<b>Statistik &amp; Ranking:</b>
• /stats — Statistik personal
• /stats @user — Statistik member lain
• /leaderboard — Top 10 (win rate)
• /leaderboard rr — Top 10 (total RR)

<b>Export &amp; Report:</b>
• /export — Export journal (CSV)
• /export pdf — Export journal (PDF)
• /report — Summary report minggu ini

<b>Lainnya:</b>
• /profile — Profil & level gamifikasi
• /help — Tampilkan pesan ini

<b>Tips:</b>
• RR bisa: +2, -1, 0 (BE), atau WIN 2RR, LOSS 1RR, BE
• Kalau RR kosong, trade dicatat sebagai OPEN
• Session: Asia, London, NY AM, NY PM`
}

// FormatWeeklySummary formats a weekly community report as HTML.
func FormatWeeklySummary(s *service.WeeklySummary) string {
	var sb strings.Builder
	sb.WriteString("📈 <b>ARK Vault — Weekly Report</b>\n")
	sb.WriteString(fmt.Sprintf("📅 %s — %s\n\n",
		s.PeriodStart.Format("02 Jan"),
		s.PeriodEnd.Format("02 Jan 2006")))

	sb.WriteString(fmt.Sprintf("Total trades: <b>%d</b>\n", s.TotalTrades))
	sb.WriteString(fmt.Sprintf("Active members: <b>%d</b>\n", s.TotalMembers))
	sb.WriteString(fmt.Sprintf("Community win rate: <b>%.1f%%</b>\n", s.CommunityWR))
	sb.WriteString(fmt.Sprintf("Total RR: <b>%+.1fR</b>\n", s.TotalRR))
	if s.MostTraded != "" {
		sb.WriteString(fmt.Sprintf("Most traded: <b>%s</b>\n", s.MostTraded))
	}

	if len(s.TopPerformers) > 0 {
		sb.WriteString("\n🏆 <b>Top Performers (by RR):</b>\n")
		medals := []string{"🥇", "🥈", "🥉"}
		for i, e := range s.TopPerformers {
			medal := fmt.Sprintf("%d.", i+1)
			if i < len(medals) {
				medal = medals[i]
			}
			name := e.Username
			if name == "" {
				name = e.FirstName
			}
			sb.WriteString(fmt.Sprintf("%s @%s — %+.1fR (%.1f%% WR)\n",
				medal, name, e.TotalRR, e.WinRate))
		}
	}

	if s.TotalTrades == 0 {
		sb.WriteString("\n📭 Belum ada trade minggu ini.")
	}

	return sb.String()
}

// FormatTradeConfirmationWithXP formats a saved trade confirmation with gamification XP info.
func FormatTradeConfirmationWithXP(t *domain.Trade, xpGained int, totalXP int, level int, title string, leveledUp bool, streak int) string {
	base := FormatTradeConfirmation(t)
	var sb strings.Builder
	sb.WriteString(base)
	sb.WriteString(fmt.Sprintf("\n⚡ +%d XP (total: %d)\n", xpGained, totalXP))
	sb.WriteString(fmt.Sprintf("📊 Level %d — %s\n", level, title))
	if streak > 0 {
		sb.WriteString(fmt.Sprintf("🔥 Streak: %d hari\n", streak))
	}
	if leveledUp {
		sb.WriteString(fmt.Sprintf("\n🎉 <b>LEVEL UP!</b> Level %d — %s\n", level, title))
	}
	return sb.String()
}

// FormatProfile formats a user's gamification profile as HTML.
func FormatProfile(profile *domain.GamificationProfile, streak *domain.StreakData) string {
	var sb strings.Builder
	sb.WriteString("🏆 <b>Profil Trader</b>\n\n")

	lvl := 1
	title := "Pemula"
	totalXP := 0
	if profile != nil {
		lvl = profile.Level
		title = profile.Title
		totalXP = profile.TotalXP
		if lvl == 0 {
			lvl = 1
			title = "Pemula"
		}
	}

	nextXP := domain.XPForNextLevel(totalXP)
	nextTitle := ""
	if nextXP > 0 {
		_, nextTitle = domain.LevelForXP(nextXP)
		sb.WriteString(fmt.Sprintf("Level: <b>%d</b> — %s\n", lvl, title))
		sb.WriteString(fmt.Sprintf("XP: <b>%d</b> / %d (next: %s)\n", totalXP, nextXP, nextTitle))

		// Progress bar
		var prevXP int
		for _, lt := range domain.LevelTable {
			if lt.Level == lvl {
				prevXP = lt.XP
				break
			}
		}
		progress := 0
		span := nextXP - prevXP
		if span > 0 {
			progress = (totalXP - prevXP) * 100 / span
		}
		filled := progress * 16 / 100
		if filled < 0 {
			filled = 0
		}
		if filled > 16 {
			filled = 16
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", 16-filled)
		sb.WriteString(fmt.Sprintf("[%s] %d%%\n", bar, progress))
	} else {
		sb.WriteString(fmt.Sprintf("Level: <b>%d</b> — %s\n", lvl, title))
		sb.WriteString(fmt.Sprintf("XP: <b>%d</b> (MAX LEVEL)\n", totalXP))
		sb.WriteString("[████████████████] 100%\n")
	}

	curStreak := 0
	longestStreak := 0
	if streak != nil {
		curStreak = streak.CurrentStreak
		longestStreak = streak.LongestStreak
	}

	sb.WriteString(fmt.Sprintf("\n🔥 Streak: <b>%d</b> hari berturut-turut\n", curStreak))
	sb.WriteString(fmt.Sprintf("📊 Streak terpanjang: <b>%d</b> hari\n", longestStreak))

	return sb.String()
}
