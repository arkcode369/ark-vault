package domain

// Stats holds computed statistics for a single member.
type Stats struct {
	TotalTrades int
	Wins        int
	Losses      int
	BreakEvens  int
	OpenTrades  int
	WinRate     float64 // percentage 0-100
	TotalRR     float64
	AvgRR       float64
	BestRR      float64
	WorstRR     float64
	CurStreak   int // positive = win streak, negative = loss streak
	MaxWinStrk  int
	ByAsset     map[AssetType]AssetStats
}

// AssetStats is a per-asset-type breakdown.
type AssetStats struct {
	Total   int
	Wins    int
	Losses  int
	TotalRR float64
	WinRate float64
}

// CalculateStats computes statistics from a slice of trades.
func CalculateStats(trades []Trade) Stats {
	s := Stats{ByAsset: make(map[AssetType]AssetStats)}

	var rrSum float64
	var rrCount int
	var curStreak, maxWin int

	bestInit := false
	for _, t := range trades {
		s.TotalTrades++

		// Per-asset accumulator
		as := s.ByAsset[t.AssetType]
		as.Total++

		switch t.Status {
		case StatusWin:
			s.Wins++
			as.Wins++
			as.TotalRR += t.ResultRR
			s.TotalRR += t.ResultRR
			if curStreak > 0 {
				curStreak++
			} else {
				curStreak = 1
			}
			if curStreak > maxWin {
				maxWin = curStreak
			}
		case StatusLoss:
			s.Losses++
			as.Losses++
			as.TotalRR += t.ResultRR // negative value
			s.TotalRR += t.ResultRR
			if curStreak < 0 {
				curStreak--
			} else {
				curStreak = -1
			}
		case StatusBE:
			s.BreakEvens++
			curStreak = 0
		case StatusOpen:
			s.OpenTrades++
		}

		if t.ResultRR != 0 {
			rrSum += t.ResultRR
			rrCount++
		}

		if !bestInit {
			s.BestRR = t.ResultRR
			s.WorstRR = t.ResultRR
			bestInit = true
		} else {
			if t.ResultRR > s.BestRR {
				s.BestRR = t.ResultRR
			}
			if t.ResultRR < s.WorstRR {
				s.WorstRR = t.ResultRR
			}
		}

		// Update per-asset win rate
		if closed := as.Wins + as.Losses; closed > 0 {
			as.WinRate = float64(as.Wins) / float64(closed) * 100
		}
		s.ByAsset[t.AssetType] = as
	}

	closed := s.Wins + s.Losses
	if closed > 0 {
		s.WinRate = float64(s.Wins) / float64(closed) * 100
	}
	if rrCount > 0 {
		s.AvgRR = rrSum / float64(rrCount)
	}
	s.CurStreak = curStreak
	s.MaxWinStrk = maxWin

	return s
}
