package domain

// Stats holds computed statistics for a single member.
type Stats struct {
	TotalTrades int
	Wins        int
	Losses      int
	BreakEvens  int
	OpenTrades  int
	WinRate     float64 // percentage 0-100
	TotalPips   float64
	AvgRR       float64
	BestPips    float64
	WorstPips   float64
	CurStreak   int // positive = win streak, negative = loss streak
	MaxWinStrk  int
	ByAsset     map[AssetType]AssetStats
}

// AssetStats is a per-asset-type breakdown.
type AssetStats struct {
	Total   int
	Wins    int
	Losses  int
	Pips    float64
	WinRate float64
}

// CalculateStats computes statistics from a slice of trades.
func CalculateStats(trades []Trade) Stats {
	s := Stats{ByAsset: make(map[AssetType]AssetStats)}

	var rrSum float64
	var rrCount int
	var curStreak, maxWin int

	for _, t := range trades {
		s.TotalTrades++

		// Per-asset accumulator
		as := s.ByAsset[t.AssetType]
		as.Total++

		switch t.Status {
		case StatusWin:
			s.Wins++
			as.Wins++
			as.Pips += t.ResultPips
			s.TotalPips += t.ResultPips
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
			as.Pips += t.ResultPips // negative value
			s.TotalPips += t.ResultPips
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

		if t.RRRatio != 0 {
			rrSum += t.RRRatio
			rrCount++
		}

		if t.ResultPips > s.BestPips {
			s.BestPips = t.ResultPips
		}
		if t.ResultPips < s.WorstPips {
			s.WorstPips = t.ResultPips
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
