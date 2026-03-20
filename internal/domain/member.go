package domain

import "time"

// Member represents a community member tracked by the bot.
type Member struct {
	TelegramID int64
	Username   string
	FirstName  string
	JoinDate   time.Time

	// Notion-specific: the page ID that belongs to this member.
	NotionPageID string
	// Notion-specific: the inline database ID inside the member page.
	NotionDBID string
}
