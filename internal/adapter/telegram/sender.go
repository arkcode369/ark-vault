package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/arkcode369/ark-vault/internal/ports"
)

// Sender implements ports.Messenger for Telegram.
type Sender struct {
	token  string
	client *http.Client
}

// NewSender creates a Telegram Sender.
func NewSender(token string) *Sender {
	return &Sender{
		token:  token,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *Sender) apiURL(method string) string {
	return fmt.Sprintf("%s%s/%s", telegramAPI, s.token, method)
}

func (s *Sender) call(ctx context.Context, method string, payload interface{}) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL(method), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w, body: %s", err, string(raw))
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram error: %s", string(raw))
	}
	return result.Result, nil
}

func extractMsgID(raw json.RawMessage) (int, error) {
	var msg struct {
		MessageID int `json:"message_id"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}

// SendText sends a plain text message.
func (s *Sender) SendText(ctx context.Context, chatID int64, text string) (int, error) {
	raw, err := s.call(ctx, "sendMessage", map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return 0, err
	}
	return extractMsgID(raw)
}

// SendHTML sends an HTML-formatted message.
func (s *Sender) SendHTML(ctx context.Context, chatID int64, html string) (int, error) {
	raw, err := s.call(ctx, "sendMessage", map[string]interface{}{
		"chat_id":    chatID,
		"text":       html,
		"parse_mode": "HTML",
	})
	if err != nil {
		return 0, err
	}
	return extractMsgID(raw)
}

// SendHTMLToThread sends an HTML-formatted message to a specific topic/thread.
// If threadID is 0, behaves like SendHTML (sends to General / main chat).
func (s *Sender) SendHTMLToThread(ctx context.Context, chatID int64, threadID int, html string) (int, error) {
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       html,
		"parse_mode": "HTML",
	}
	if threadID > 0 {
		payload["message_thread_id"] = threadID
	}
	raw, err := s.call(ctx, "sendMessage", payload)
	if err != nil {
		return 0, err
	}
	return extractMsgID(raw)
}

// SendWithKeyboard sends a message with inline keyboard.
func (s *Sender) SendWithKeyboard(ctx context.Context, chatID int64, text string, rows [][]ports.InlineButton) (int, error) {
	raw, err := s.call(ctx, "sendMessage", map[string]interface{}{
		"chat_id":      chatID,
		"text":         text,
		"parse_mode":   "HTML",
		"reply_markup": buildKeyboard(rows),
	})
	if err != nil {
		return 0, err
	}
	return extractMsgID(raw)
}

// EditMessage edits an existing message's text.
func (s *Sender) EditMessage(ctx context.Context, chatID int64, msgID int, text string) error {
	_, err := s.call(ctx, "editMessageText", map[string]interface{}{
		"chat_id":    chatID,
		"message_id": msgID,
		"text":       text,
		"parse_mode": "HTML",
	})
	return err
}

// EditWithKeyboard edits text and keyboard.
func (s *Sender) EditWithKeyboard(ctx context.Context, chatID int64, msgID int, text string, rows [][]ports.InlineButton) error {
	_, err := s.call(ctx, "editMessageText", map[string]interface{}{
		"chat_id":      chatID,
		"message_id":   msgID,
		"text":         text,
		"parse_mode":   "HTML",
		"reply_markup": buildKeyboard(rows),
	})
	return err
}

// AnswerCallback acknowledges a callback query.
func (s *Sender) AnswerCallback(ctx context.Context, callbackID string, text string) error {
	_, err := s.call(ctx, "answerCallbackQuery", map[string]interface{}{
		"callback_query_id": callbackID,
		"text":              text,
	})
	return err
}

// DeleteMessage removes a message.
func (s *Sender) DeleteMessage(ctx context.Context, chatID int64, msgID int) error {
	_, err := s.call(ctx, "deleteMessage", map[string]interface{}{
		"chat_id":    chatID,
		"message_id": msgID,
	})
	return err
}

// SendDocument sends a file to the chat.
func (s *Sender) SendDocument(ctx context.Context, chatID int64, filename string, data []byte, caption string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("chat_id", fmt.Sprintf("%d", chatID))
	if caption != "" {
		_ = w.WriteField("caption", caption)
	}
	part, err := w.CreateFormFile("document", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL("sendDocument"), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func buildKeyboard(rows [][]ports.InlineButton) map[string]interface{} {
	tgRows := make([][]map[string]interface{}, len(rows))
	for i, row := range rows {
		tgRows[i] = make([]map[string]interface{}, len(row))
		for j, btn := range row {
			tgRows[i][j] = map[string]interface{}{
				"text":          btn.Text,
				"callback_data": btn.CallbackData,
			}
		}
	}
	return map[string]interface{}{
		"inline_keyboard": tgRows,
	}
}
