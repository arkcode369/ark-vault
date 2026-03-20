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
	sb.WriteString(fmt.Sprintf("Total Pips: <b>%+.1f</b>\n", s.TotalPips))
	sb.WriteString(fmt.Sprintf("Best: <b>%+.1f</b> pips\n", s.BestPips))
	sb.WriteString(fmt.Sprintf("Worst: <b>%+.1f</b> pips\n", s.WorstPips))

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
			sb.WriteString(fmt.Sprintf("  %s: %d trades, %.1f%% WR, %+.1f pips\n",
				at.String(), as.Total, as.WinRate, as.Pips))
		}
	}

	return sb.String()
}

// FormatLeaderboard formats the leaderboard as an HTML Telegram message.
func FormatLeaderboard(entries []service.LeaderboardEntry, metric string) string {
	var sb strings.Builder
	sb.WriteString("🏆 <b>Leaderboard</b>")
	if metric == "pips" {
		sb.WriteString(" (by Pips)")
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
		if metric == "pips" {
			sb.WriteString(fmt.Sprintf("%s <b>@%s</b> — %+.1f pips (%d trades)\n",
				rank, name, e.TotalPips, e.TotalTrades))
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
	sb.WriteString(fmt.Sprintf("Entry: <b>%g</b>\n", t.EntryPrice))
	sb.WriteString(fmt.Sprintf("SL: <b>%g</b> | TP: <b>%g</b>\n", t.StopLoss, t.TakeProfit))
	sb.WriteString(fmt.Sprintf("Status: <b>%s</b>", t.Status))
	if t.ResultPips != 0 {
		sb.WriteString(fmt.Sprintf(" (%+.1f pips)", t.ResultPips))
	}
	sb.WriteString("\n")
	return sb.String()
}

// FormatHelp returns the /help message.
func FormatHelp() string {
	return `🔐 <b>ARK Vault — Help</b>

<b>Mencatat Trade:</b>
• <code>/journal</code> — Guided flow (step-by-step dengan tombol)
• Kirim pesan dengan format:
<pre>#journal
Pair: EURUSD
Type: BUY
Entry: 1.0850
SL: 1.0800
TP: 1.0950
Result: WIN +50 pips</pre>
  Bisa attach screenshot bersamaan.

<b>Statistik &amp; Ranking:</b>
• <code>/stats</code> — Statistik personal
• <code>/stats @user</code> — Statistik member lain
• <code>/leaderboard</code> — Top 10 (win rate)
• <code>/leaderboard pips</code> — Top 10 (total pips)

<b>Trade Management:</b>
• <code>/close [id] [price] [pips] [WIN/LOSS/BE]</code> — Tutup trade

<b>Export &amp; Report:</b>
• <code>/export</code> — Export journal (CSV)
• <code>/export pdf</code> — Export journal (PDF)
• <code>/report</code> — Summary report minggu ini

<b>Lainnya:</b>
• <code>/help</code> — Tampilkan pesan ini

<b>Tips:</b>
• Result bisa: WIN +50 pips, LOSS -30 pips, atau BE
• Kalau Result kosong, trade dicatat sebagai OPEN`
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
	sb.WriteString(fmt.Sprintf("Total pips: <b>%+.1f</b>\n", s.TotalPips))
	if s.MostTraded != "" {
		sb.WriteString(fmt.Sprintf("Most traded: <b>%s</b>\n", s.MostTraded))
	}

	if len(s.TopPerformers) > 0 {
		sb.WriteString("\n🏆 <b>Top Performers (by Pips):</b>\n")
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
			sb.WriteString(fmt.Sprintf("%s @%s — %+.1f pips (%.1f%% WR)\n",
				medal, name, e.TotalPips, e.WinRate))
		}
	}

	if s.TotalTrades == 0 {
		sb.WriteString("\n📭 Belum ada trade minggu ini.")
	}

	return sb.String()
}
