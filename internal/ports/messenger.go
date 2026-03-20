package ports

import "context"

// InlineButton is a single button in an inline keyboard.
type InlineButton struct {
	Text         string
	CallbackData string
}

// Messenger abstracts the messaging platform (Telegram).
type Messenger interface {
	// SendText sends a plain-text message and returns the message ID.
	SendText(ctx context.Context, chatID int64, text string, threadID ...int) (int, error)

	// SendHTML sends an HTML-formatted message.
	SendHTML(ctx context.Context, chatID int64, html string, threadID ...int) (int, error)

	// SendWithKeyboard sends a message with inline keyboard buttons.
	// rows is a 2D slice: each inner slice is one row of buttons.
	SendWithKeyboard(ctx context.Context, chatID int64, text string, rows [][]InlineButton, threadID ...int) (int, error)

	// EditMessage replaces the text of an existing message.
	EditMessage(ctx context.Context, chatID int64, msgID int, text string) error

	// EditWithKeyboard replaces text and keyboard of an existing message.
	EditWithKeyboard(ctx context.Context, chatID int64, msgID int, text string, rows [][]InlineButton) error

	// AnswerCallback acknowledges a callback query with optional toast text.
	AnswerCallback(ctx context.Context, callbackID string, text string) error

	// DeleteMessage removes a message.
	DeleteMessage(ctx context.Context, chatID int64, msgID int) error

	// SendDocument sends a file (e.g. CSV export) to the chat.
	SendDocument(ctx context.Context, chatID int64, filename string, data []byte, caption string, threadID ...int) error
}
