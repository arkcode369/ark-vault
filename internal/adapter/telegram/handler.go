package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/arkcode369/ark-vault/internal/domain"
	"github.com/arkcode369/ark-vault/internal/ports"
	"github.com/arkcode369/ark-vault/internal/service"
	"github.com/arkcode369/ark-vault/pkg/timeutil"
)

// Handler routes incoming Telegram updates to the appropriate logic.
type Handler struct {
	sender           *Sender
	bot              *Bot
	journal          *service.JournalService
	leaderboard      *service.LeaderboardService
	report           *service.ReportService
	exporter         ports.Exporter
	trades           ports.TradeRepository
	members          ports.MemberRepository
	guided           *GuidedFlow
	limiter          *RateLimiter
	logger           *slog.Logger
	communityGroupID int64
	ownerID          int64
}

// NewHandler creates a Handler.
func NewHandler(
	sender *Sender,
	journal *service.JournalService,
	leaderboard *service.LeaderboardService,
	report *service.ReportService,
	exporter ports.Exporter,
	trades ports.TradeRepository,
	members ports.MemberRepository,
	limiter *RateLimiter,
	logger *slog.Logger,
	communityGroupID int64,
	ownerID int64,
) *Handler {
	return &Handler{
		sender:           sender,
		journal:          journal,
		leaderboard:      leaderboard,
		report:           report,
		exporter:         exporter,
		trades:           trades,
		members:          members,
		guided:           NewGuidedFlow(),
		limiter:          limiter,
		logger:           logger,
		communityGroupID: communityGroupID,
		ownerID:          ownerID,
	}
}

// SetBot sets the Bot reference (needed for file downloads).
func (h *Handler) SetBot(b *Bot) {
	h.bot = b
}

// isAuthorized checks if the user is a member of the community group.
// If communityGroupID is 0, everyone is authorized.
// In group chats, non-members are silently ignored (no spam in group).
// In private chats, non-members get a friendly denial with owner contact.
func (h *Handler) isAuthorized(ctx context.Context, userID int64, chatID int64, threadID int, chatType string) bool {
	if h.communityGroupID == 0 || h.bot == nil {
		return true // no community gate configured
	}
	isMember, err := h.bot.CheckChatMember(ctx, h.communityGroupID, userID)
	if err != nil {
		h.logger.Error("check membership failed", "error", err, "user_id", userID)
		return true // fail open on API errors
	}
	if !isMember {
		// Only send denial message in private chat — don't spam group chats
		if chatType == "private" {
			msg := "🔒 Fitur ini hanya tersedia untuk member komunitas."
			if h.ownerID > 0 {
				msg += fmt.Sprintf("\n\nUntuk bergabung, hubungi owner:\n➡ <a href=\"tg://user?id=%d\">Contact Owner</a>", h.ownerID)
			} else {
				msg += "\n\nHubungi admin komunitas untuk bergabung."
			}
			h.sender.SendHTML(ctx, chatID, msg, threadID)
		}
		// Notify owner about unauthorized access attempt
		if h.ownerID > 0 {
			notif := fmt.Sprintf("⚠️ <b>Unauthorized access attempt</b>\n\nUser: <a href=\"tg://user?id=%d\">%d</a>\nChat: <code>%d</code>\nType: %s",
				userID, userID, chatID, chatType)
			h.sender.SendHTML(ctx, h.ownerID, notif)
		}
		return false
	}
	return true
}

// HandleUpdate processes a single Telegram update.
func (h *Handler) HandleUpdate(ctx context.Context, u Update) {
	if u.CallbackQuery != nil {
		h.handleCallback(ctx, u.CallbackQuery)
		return
	}
	if u.Message == nil {
		return
	}

	msg := u.Message
	text := strings.TrimSpace(msg.Text)
	if text == "" && msg.Caption != "" {
		text = strings.TrimSpace(msg.Caption)
	}

	// Rate limiting (skip for #journal text which is passive)
	if msg.From != nil && strings.HasPrefix(text, "/") {
		if !h.limiter.Allow(msg.From.ID) {
			h.sender.SendText(ctx, msg.Chat.ID, "⏳ Terlalu banyak request. Coba lagi nanti.", msg.MessageThreadID)
			return
		}
	}

	// Community membership gate
	if msg.From != nil && !h.isAuthorized(ctx, msg.From.ID, msg.Chat.ID, msg.MessageThreadID, msg.Chat.Type) {
		return
	}

	// Check if user is in a guided session and this is a free-text input step
	if msg.From != nil {
		if session := h.guided.Get(msg.From.ID); session != nil {
			h.handleGuidedInput(ctx, msg, session)
			return
		}
	}

	// Command routing
	switch {
	case strings.HasPrefix(text, "/journal"):
		h.cmdJournal(ctx, msg)
	case strings.HasPrefix(text, "/stats"):
		h.cmdStats(ctx, msg, text)
	case strings.HasPrefix(text, "/leaderboard"):
		h.cmdLeaderboard(ctx, msg, text)
	case strings.HasPrefix(text, "/export"):
		h.cmdExport(ctx, msg, text)
	case strings.HasPrefix(text, "/report"):
		h.cmdReport(ctx, msg)
	case strings.HasPrefix(text, "/help"), strings.HasPrefix(text, "/start"):
		h.sender.SendHTML(ctx, msg.Chat.ID, FormatHelp(), msg.MessageThreadID)
	case strings.HasPrefix(text, "#journal"):
		h.handleTextJournal(ctx, msg, text)
	}
}

// handleTextJournal processes a #journal text-format trade entry.
func (h *Handler) handleTextJournal(ctx context.Context, msg *Message, text string) {
	result := ParseJournalMessage(text)
	if result.Err != "" {
		h.sender.SendHTML(ctx, msg.Chat.ID, "❌ "+result.Err, msg.MessageThreadID)
		return
	}

	trade := result.Trade
	username := ""
	firstName := ""
	if msg.From != nil {
		username = msg.From.Username
		firstName = msg.From.FirstName
	}

	if err := h.journal.RecordTrade(ctx, msg.From.ID, username, firstName, trade); err != nil {
		h.logger.Error("record trade failed", "error", err)
		h.sender.SendText(ctx, msg.Chat.ID, "❌ Gagal menyimpan trade: "+err.Error(), msg.MessageThreadID)
		return
	}

	// Handle attached photo
	if len(msg.Photo) > 0 && trade.ID != "" && h.bot != nil {
		h.uploadPhoto(ctx, msg.Photo, trade.ID)
	}

	h.sender.SendHTML(ctx, msg.Chat.ID, FormatTradeConfirmation(trade), msg.MessageThreadID)
}

// cmdJournal starts the guided flow.
func (h *Handler) cmdJournal(ctx context.Context, msg *Message) {
	if msg.From == nil {
		return
	}
	session := h.guided.Start(msg.From.ID, msg.Chat.ID, msg.MessageThreadID)
	text, btns := StepPrompt(session)
	rows := convertButtons(btns)
	msgID, err := h.sender.SendWithKeyboard(ctx, msg.Chat.ID, text, rows, msg.MessageThreadID)
	if err != nil {
		h.logger.Error("send guided prompt", "error", err)
		return
	}
	session.MsgID = msgID
}

// cmdStats shows personal or another member's stats.
func (h *Handler) cmdStats(ctx context.Context, msg *Message, text string) {
	if msg.From == nil {
		return
	}
	targetID := msg.From.ID
	targetUsername := msg.From.Username

	// Check if querying another user: /stats @username
	parts := strings.Fields(text)
	if len(parts) > 1 && strings.HasPrefix(parts[1], "@") {
		lookupUsername := strings.TrimPrefix(parts[1], "@")
		member, err := h.findMemberByUsername(ctx, lookupUsername)
		if err != nil || member == nil {
			h.sender.SendText(ctx, msg.Chat.ID, fmt.Sprintf("📭 Member @%s belum pernah mencatat trade.", lookupUsername), msg.MessageThreadID)
			return
		}
		targetID = member.TelegramID
		targetUsername = member.Username
	}

	stats, err := h.journal.GetMemberStats(ctx, targetID)
	if err != nil {
		h.sender.SendText(ctx, msg.Chat.ID, "📭 Belum ada trade yang tercatat.", msg.MessageThreadID)
		return
	}
	if stats == nil {
		h.sender.SendText(ctx, msg.Chat.ID, "📭 Belum ada trade yang tercatat.", msg.MessageThreadID)
		return
	}

	h.sender.SendHTML(ctx, msg.Chat.ID, FormatStats(targetUsername, stats), msg.MessageThreadID)
}

// findMemberByUsername searches cached members by username.
func (h *Handler) findMemberByUsername(ctx context.Context, username string) (*domain.Member, error) {
	members, err := h.members.ListMembers(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		if strings.EqualFold(m.Username, username) {
			return &m, nil
		}
	}
	return nil, nil
}

// cmdLeaderboard shows the leaderboard.
func (h *Handler) cmdLeaderboard(ctx context.Context, msg *Message, text string) {
	metric := "winrate"
	if strings.Contains(strings.ToLower(text), "rr") {
		metric = "rr"
	}

	entries, err := h.leaderboard.GetLeaderboard(ctx, metric, 10, 5)
	if err != nil {
		h.sender.SendText(ctx, msg.Chat.ID, "❌ Error: "+err.Error(), msg.MessageThreadID)
		return
	}

	h.sender.SendHTML(ctx, msg.Chat.ID, FormatLeaderboard(entries, metric), msg.MessageThreadID)
}

// cmdExport exports the member's trades as CSV or PDF.
func (h *Handler) cmdExport(ctx context.Context, msg *Message, text string) {
	if msg.From == nil {
		return
	}
	trades, err := h.trades.GetTrades(ctx, msg.From.ID)
	if err != nil {
		h.sender.SendText(ctx, msg.Chat.ID, "📭 Belum ada trade untuk di-export.", msg.MessageThreadID)
		return
	}
	if len(trades) == 0 {
		h.sender.SendText(ctx, msg.Chat.ID, "📭 Belum ada trade untuk di-export.", msg.MessageThreadID)
		return
	}

	if h.exporter == nil {
		h.sender.SendText(ctx, msg.Chat.ID, "⚠️ Export belum tersedia.", msg.MessageThreadID)
		return
	}

	// Check if PDF requested: /export pdf
	format := "csv"
	parts := strings.Fields(text)
	if len(parts) > 1 && strings.EqualFold(parts[1], "pdf") {
		format = "pdf"
	}

	username := msg.From.Username
	if username == "" {
		username = msg.From.FirstName
	}

	switch format {
	case "pdf":
		stats, _ := h.journal.GetMemberStats(ctx, msg.From.ID)
		data, err := h.exporter.ExportPDF(ctx, username, trades, stats)
		if err != nil {
			h.sender.SendText(ctx, msg.Chat.ID, "❌ Export PDF gagal: "+err.Error(), msg.MessageThreadID)
			return
		}
		filename := fmt.Sprintf("journal_%s.pdf", username)
		h.sender.SendDocument(ctx, msg.Chat.ID, filename, data, "📊 Export trade journal (PDF)", msg.MessageThreadID)
	default:
		data, err := h.exporter.ExportCSV(ctx, trades)
		if err != nil {
			h.sender.SendText(ctx, msg.Chat.ID, "❌ Export CSV gagal: "+err.Error(), msg.MessageThreadID)
			return
		}
		filename := fmt.Sprintf("journal_%s.csv", username)
		h.sender.SendDocument(ctx, msg.Chat.ID, filename, data, "📊 Export trade journal (CSV)", msg.MessageThreadID)
	}
}

// cmdReport generates and sends a weekly community summary.
func (h *Handler) cmdReport(ctx context.Context, msg *Message) {
	now := time.Now().UTC()
	weekStart := timeutil.StartOfWeek(now, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	summary, err := h.report.GenerateWeeklySummary(ctx, weekStart, weekEnd)
	if err != nil {
		h.sender.SendText(ctx, msg.Chat.ID, "❌ Gagal generate report: "+err.Error(), msg.MessageThreadID)
		return
	}

	h.sender.SendHTML(ctx, msg.Chat.ID, FormatWeeklySummary(summary), msg.MessageThreadID)
}

// SendScheduledReport sends the weekly report to a specific chat and optional topic thread.
// Called by the scheduler.
func (h *Handler) SendScheduledReport(ctx context.Context, chatID int64, threadID int) error {
	now := time.Now().UTC()
	weekStart := timeutil.StartOfWeek(now, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)

	summary, err := h.report.GenerateWeeklySummary(ctx, weekStart, weekEnd)
	if err != nil {
		return err
	}

	_, err = h.sender.SendHTMLToThread(ctx, chatID, threadID, FormatWeeklySummary(summary))
	return err
}

// handleCallback processes inline button presses.
func (h *Handler) handleCallback(ctx context.Context, cb *CallbackQuery) {
	h.sender.AnswerCallback(ctx, cb.ID, "")

	// Community gate
	if cb.Message != nil && !h.isAuthorized(ctx, cb.From.ID, cb.Message.Chat.ID, cb.Message.MessageThreadID, cb.Message.Chat.Type) {
		return
	}

	session := h.guided.Get(cb.From.ID)
	if session == nil {
		return
	}

	data := cb.Data

	switch {
	case strings.HasPrefix(data, "asset:"):
		assetStr := strings.TrimPrefix(data, "asset:")
		session.Trade.AssetType = domain.AssetType(assetStr)
		session.Step = StepSymbol

	case strings.HasPrefix(data, "dir:"):
		dirStr := strings.TrimPrefix(data, "dir:")
		session.Trade.Direction = domain.Direction(dirStr)
		session.Step = StepResult

	case strings.HasPrefix(data, "result:"):
		resultStr := strings.TrimPrefix(data, "result:")
		switch resultStr {
		case "WIN":
			session.Trade.Status = domain.StatusWin
			session.Step = StepRRAmount
		case "LOSS":
			session.Trade.Status = domain.StatusLoss
			session.Step = StepRRAmount
		case "BE":
			session.Trade.Status = domain.StatusBE
			session.Trade.ResultRR = 0
			session.Step = StepTimeWindow
		case "OPEN":
			session.Trade.Status = domain.StatusOpen
			session.Step = StepTimeWindow
		}

	case strings.HasPrefix(data, "session:"):
		sessionStr := strings.TrimPrefix(data, "session:")
		switch sessionStr {
		case "asia":
			session.Trade.TimeWindow = domain.SessionAsia
		case "london":
			session.Trade.TimeWindow = domain.SessionLondon
		case "nyam":
			session.Trade.TimeWindow = domain.SessionNYAM
		case "nypm":
			session.Trade.TimeWindow = domain.SessionNYPM
		case "skip":
			// leave empty
		}
		session.Step = StepConfluence

	case data == "screenshot:skip":
		session.Step = StepConfirm

	case data == "confirm:yes":
		h.submitGuidedTrade(ctx, cb.From, session)
		return

	case data == "confirm:no":
		h.guided.Remove(cb.From.ID)
		if cb.Message != nil {
			h.sender.EditMessage(ctx, cb.Message.Chat.ID, cb.Message.MessageID, "❌ Trade dibatalkan.")
		}
		return
	}

	// Update the prompt message
	text, btns := StepPrompt(session)
	rows := convertButtons(btns)
	if cb.Message != nil {
		if len(rows) > 0 {
			h.sender.EditWithKeyboard(ctx, cb.Message.Chat.ID, cb.Message.MessageID, text, rows)
		} else {
			h.sender.EditMessage(ctx, cb.Message.Chat.ID, cb.Message.MessageID, text)
		}
		session.MsgID = cb.Message.MessageID
	}
}

// handleGuidedInput processes free-text input during a guided flow.
func (h *Handler) handleGuidedInput(ctx context.Context, msg *Message, session *GuidedSession) {
	text := strings.TrimSpace(msg.Text)

	switch session.Step {
	case StepSymbol:
		session.Trade.Symbol = strings.ToUpper(strings.ReplaceAll(text, "/", ""))
		session.Step = StepDirection

	case StepRRAmount:
		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			h.sender.SendText(ctx, msg.Chat.ID, "❌ Masukkan angka yang valid untuk RR (contoh: 2, 1.5).", session.ThreadID)
			return
		}
		if session.Trade.Status == domain.StatusWin {
			if v < 0 {
				v = -v
			}
			session.Trade.ResultRR = v
		} else if session.Trade.Status == domain.StatusLoss {
			if v > 0 {
				v = -v
			}
			session.Trade.ResultRR = v
		}
		session.Step = StepTimeWindow

	case StepConfluence:
		if strings.ToLower(text) != "skip" {
			session.Trade.Confluence = text
		}
		session.Step = StepScreenshot

	case StepScreenshot:
		// Check if it's a photo message
		if len(msg.Photo) > 0 {
			// Get largest photo
			photo := msg.Photo[len(msg.Photo)-1]
			session.PhotoFileID = photo.FileID
			session.Step = StepConfirm
		} else {
			h.sender.SendText(ctx, msg.Chat.ID, "❌ Kirim foto, atau tekan Skip.", session.ThreadID)
			return
		}

	default:
		return
	}

	// Send next step prompt
	promptText, btns := StepPrompt(session)
	rows := convertButtons(btns)
	if len(rows) > 0 {
		msgID, _ := h.sender.SendWithKeyboard(ctx, msg.Chat.ID, promptText, rows, session.ThreadID)
		session.MsgID = msgID
	} else {
		msgID, _ := h.sender.SendHTML(ctx, msg.Chat.ID, promptText, session.ThreadID)
		session.MsgID = msgID
	}
}

// submitGuidedTrade finalises the guided flow and saves the trade.
func (h *Handler) submitGuidedTrade(ctx context.Context, from *User, session *GuidedSession) {
	if session.Submitting {
		return // already processing
	}
	session.Submitting = true

	trade := &session.Trade
	if trade.Status == "" {
		trade.Status = domain.StatusOpen
	}

	err := h.journal.RecordTrade(ctx, from.ID, from.Username, from.FirstName, trade)
	if err != nil {
		h.logger.Error("guided submit failed", "error", err)
		h.sender.SendText(ctx, session.ChatID, "❌ Gagal menyimpan: "+err.Error(), session.ThreadID)
		h.guided.Remove(from.ID)
		return
	}

	// Upload screenshot if provided
	if session.PhotoFileID != "" && trade.ID != "" && h.bot != nil {
		h.uploadPhoto(ctx, []PhotoSize{{FileID: session.PhotoFileID}}, trade.ID)
	}

	h.sender.SendHTML(ctx, session.ChatID, FormatTradeConfirmation(trade), session.ThreadID)
	h.guided.Remove(from.ID)
}

// uploadPhoto downloads a photo from Telegram and attaches it to the trade's Notion page.
func (h *Handler) uploadPhoto(ctx context.Context, photos []PhotoSize, tradeID string) {
	if len(photos) == 0 || h.bot == nil {
		return
	}
	// Use the largest photo
	photo := photos[len(photos)-1]
	fileURL, err := h.bot.GetFileURL(ctx, photo.FileID)
	if err != nil {
		h.logger.Error("get file url", "error", err)
		return
	}
	// Pass the URL as the "filename" — the Notion image store uses external URLs
	if err := h.journal.UploadScreenshot(ctx, tradeID, fileURL, nil); err != nil {
		h.logger.Error("upload screenshot", "error", err)
	}
}

// convertButtons converts internal button types to port types.
func convertButtons(btns [][]InlineBtn) [][]ports.InlineButton {
	if btns == nil {
		return nil
	}
	rows := make([][]ports.InlineButton, len(btns))
	for i, row := range btns {
		rows[i] = make([]ports.InlineButton, len(row))
		for j, btn := range row {
			rows[i][j] = ports.InlineButton{
				Text:         btn.Text,
				CallbackData: btn.Data,
			}
		}
	}
	return rows
}
