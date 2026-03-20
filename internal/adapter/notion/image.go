package notion

import (
	"context"
	"fmt"
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
// For Notion, we use an external URL (Telegram file URL) since Notion API
// does not support direct file upload to page blocks. The bot downloads
// the photo from Telegram and provides the public file URL.
func (r *ImageRepo) Upload(ctx context.Context, pageID string, filename string, data []byte) (string, error) {
	// Notion API doesn't support direct binary uploads to blocks.
	// We store the Telegram file_path URL as an external image block.
	// The caller should pass the Telegram file URL as "filename" param.
	imageURL := filename

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
