package ports

import "context"

// ImageStore handles uploading trade screenshots.
type ImageStore interface {
	// Upload stores the image and returns a URL or identifier.
	// pageID is the Notion page (trade entry) to attach the image to.
	Upload(ctx context.Context, pageID string, filename string, data []byte) (string, error)
}
