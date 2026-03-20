package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const telegramAPI = "https://api.telegram.org/bot"

// Bot manages the Telegram bot lifecycle (long polling).
type Bot struct {
	token   string
	handler *Handler
	client  *http.Client
	logger  *slog.Logger
	offset  int
}

// NewBot creates a new Bot.
func NewBot(token string, handler *Handler, logger *slog.Logger) *Bot {
	return &Bot{
		token:   token,
		handler: handler,
		client:  &http.Client{Timeout: 60 * time.Second},
		logger:  logger,
	}
}

// Start begins the long-polling loop. Blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("starting telegram bot polling")
	for {
		select {
		case <-ctx.Done():
			b.logger.Info("bot polling stopped")
			return ctx.Err()
		default:
			updates, err := b.getUpdates(ctx)
			if err != nil {
				b.logger.Error("get updates failed", "error", err)
				time.Sleep(2 * time.Second)
				continue
			}
			for _, u := range updates {
				if u.UpdateID >= b.offset {
					b.offset = u.UpdateID + 1
				}
				go b.handler.HandleUpdate(ctx, u)
			}
		}
	}
}

// Update represents a Telegram update.
type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// Message represents a Telegram message.
type Message struct {
	MessageID int    `json:"message_id"`
	From      *User  `json:"from,omitempty"`
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text,omitempty"`
	Photo     []PhotoSize `json:"photo,omitempty"`
	Caption   string `json:"caption,omitempty"`
}

// User represents a Telegram user.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
}

// Chat represents a Telegram chat.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"` // "private", "group", "supergroup"
}

// PhotoSize represents one size of a photo.
type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size,omitempty"`
}

// CallbackQuery represents an inline button press.
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data,omitempty"`
}

func (b *Bot) getUpdates(ctx context.Context) ([]Update, error) {
	url := fmt.Sprintf("%s%s/getUpdates?offset=%d&timeout=30&allowed_updates=[\"message\",\"callback_query\"]",
		telegramAPI, b.token, b.offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if !result.OK {
		return nil, fmt.Errorf("telegram API error: %s", string(body))
	}
	return result.Result, nil
}

// GetFileURL retrieves the direct download URL for a file.
func (b *Bot) GetFileURL(ctx context.Context, fileID string) (string, error) {
	url := fmt.Sprintf("%s%s/getFile?file_id=%s", telegramAPI, b.token, fileID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.token, result.Result.FilePath), nil
}
