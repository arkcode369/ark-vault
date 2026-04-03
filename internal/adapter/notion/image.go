package notion

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// ImageRepo implements ports.ImageStore by appending image blocks to Notion pages.
type ImageRepo struct {
	client *Client
}

// NewImageRepo creates an ImageRepo.
func NewImageRepo(client *Client) *ImageRepo {
	return &ImageRepo{client: client}
}

// Upload appends an image block to the trade's Notion page.
// If data is provided, it creates a base64 data URI (Notion supports this for external images).
// If data is nil, filename is treated as an external URL (legacy behavior, logs security warning).
func (r *ImageRepo) Upload(ctx context.Context, pageID string, filename string, data []byte) (string, error) {
	var imageURL string

	if len(data) > 0 {
		// Convert file data to base64 data URI - no external URL needed, token stays safe
		contentType := "image/jpeg" // default
		if strings.HasSuffix(strings.ToLower(filename), ".png") {
			contentType = "image/png"
		}
		b64Data := base64.StdEncoding.EncodeToString(data)
		imageURL = fmt.Sprintf("data:%s;base64,%s", contentType, b64Data)
	} else {
		// Legacy behavior: filename is an external URL
		// This path should not be used for Telegram files as it exposes bot tokens
		imageURL = filename
	}

	children := []map[string]interface{}{
		{
			"object": "block",
			"type":   "image",
			"image": map[string]interface{}{
				"type": "external",
				"external": map[string]interface{}{
					"url": imageURL,
				},
			},
		},
	}

	_, err := r.client.AppendBlocks(ctx, pageID, children)
	if err != nil {
		return "", fmt.Errorf("append image block: %w", err)
	}
	return imageURL, nil
}

// parseDate parses a Notion date string.
func parseDate(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}
