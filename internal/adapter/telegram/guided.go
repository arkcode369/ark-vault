package telegram

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// GuidedStep tracks which step a user is on in the guided journal flow.
type GuidedStep int

const (
	StepNone GuidedStep = iota
	StepAssetType
	StepSymbol
	StepDirection
	StepEntry
	StepStopLoss
	StepTakeProfit
	StepScreenshot
	StepConfirm
)

// GuidedSession holds the in-progress state for one user's guided flow.
type GuidedSession struct {
	Step       GuidedStep
	Trade      domain.Trade
	ChatID     int64
	MsgID      int // last bot message ID, for editing
	PhotoURL   string
	PhotoFileID string
	ExpiresAt  time.Time
}

// GuidedFlow manages guided journal sessions per user.
type GuidedFlow struct {
	mu       sync.RWMutex
	sessions map[int64]*GuidedSession // keyed by Telegram user ID
	ttl      time.Duration
}

// NewGuidedFlow creates a new GuidedFlow.
func NewGuidedFlow() *GuidedFlow {
	gf := &GuidedFlow{
		sessions: make(map[int64]*GuidedSession),
		ttl:      5 * time.Minute,
	}
	go gf.cleanupLoop()
	return gf
}

// Start creates a new session for the user.
func (gf *GuidedFlow) Start(userID int64, chatID int64) *GuidedSession {
	gf.mu.Lock()
	defer gf.mu.Unlock()
	s := &GuidedSession{
		Step:      StepAssetType,
		ChatID:    chatID,
		ExpiresAt: time.Now().Add(gf.ttl),
	}
	gf.sessions[userID] = s
	return s
}

// Get returns the active session for a user, or nil.
func (gf *GuidedFlow) Get(userID int64) *GuidedSession {
	gf.mu.RLock()
	defer gf.mu.RUnlock()
	s, ok := gf.sessions[userID]
	if !ok || time.Now().After(s.ExpiresAt) {
		return nil
	}
	return s
}

// Remove deletes a user's session.
func (gf *GuidedFlow) Remove(userID int64) {
	gf.mu.Lock()
	defer gf.mu.Unlock()
	delete(gf.sessions, userID)
}

// StepPrompt returns the text and optional keyboard rows for the current step.
func StepPrompt(s *GuidedSession) (string, [][]InlineBtn) {
	switch s.Step {
	case StepAssetType:
		return "📊 <b>Step 1/7</b> — Pilih asset type:", [][]InlineBtn{
			{
				{Text: "Forex", Data: "asset:FOREX"},
				{Text: "Gold", Data: "asset:GOLD"},
			},
			{
				{Text: "Indices", Data: "asset:INDICES"},
				{Text: "Crypto", Data: "asset:CRYPTO"},
			},
		}
	case StepSymbol:
		return fmt.Sprintf("📊 <b>Step 2/7</b> — Masukkan pair/symbol:\n(Asset: %s)", s.Trade.AssetType.String()), nil
	case StepDirection:
		return "📊 <b>Step 3/7</b> — Direction:", [][]InlineBtn{
			{
				{Text: "🟢 BUY", Data: "dir:BUY"},
				{Text: "🔴 SELL", Data: "dir:SELL"},
			},
		}
	case StepEntry:
		return "📊 <b>Step 4/7</b> — Entry price:", nil
	case StepStopLoss:
		return "📊 <b>Step 5/7</b> — Stop Loss:", nil
	case StepTakeProfit:
		return "📊 <b>Step 6/7</b> — Take Profit:", nil
	case StepScreenshot:
		return "📊 <b>Step 7/7</b> — Kirim screenshot trade (opsional):", [][]InlineBtn{
			{{Text: "⏭ Skip", Data: "screenshot:skip"}},
		}
	case StepConfirm:
		return buildConfirmText(s), [][]InlineBtn{
			{
				{Text: "✅ Submit", Data: "confirm:yes"},
				{Text: "❌ Cancel", Data: "confirm:no"},
			},
		}
	default:
		return "", nil
	}
}

// InlineBtn is a shorthand for guided flow buttons (not the port type).
type InlineBtn struct {
	Text string
	Data string
}

func buildConfirmText(s *GuidedSession) string {
	t := &s.Trade
	var sb strings.Builder
	sb.WriteString("📋 <b>Konfirmasi Trade:</b>\n\n")
	sb.WriteString(fmt.Sprintf("Asset: <b>%s</b>\n", t.AssetType.String()))
	sb.WriteString(fmt.Sprintf("Symbol: <b>%s</b>\n", t.Symbol))
	sb.WriteString(fmt.Sprintf("Direction: <b>%s</b>\n", t.Direction))
	sb.WriteString(fmt.Sprintf("Entry: <b>%g</b>\n", t.EntryPrice))
	sb.WriteString(fmt.Sprintf("SL: <b>%g</b>\n", t.StopLoss))
	sb.WriteString(fmt.Sprintf("TP: <b>%g</b>\n", t.TakeProfit))
	if s.PhotoURL != "" {
		sb.WriteString("Screenshot: ✅ attached\n")
	}
	sb.WriteString("\nKirim submit untuk menyimpan.")
	return sb.String()
}

func (gf *GuidedFlow) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		gf.mu.Lock()
		now := time.Now()
		for uid, s := range gf.sessions {
			if now.After(s.ExpiresAt) {
				delete(gf.sessions, uid)
			}
		}
		gf.mu.Unlock()
	}
}
