package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL        = "https://api.notion.com/v1"
	notionVersion  = "2022-06-28"
	requestTimeout = 30 * time.Second
)

// Client is a thin wrapper around the Notion API.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a Notion API client.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// do executes an HTTP request against the Notion API.
func (c *Client) do(ctx context.Context, method, path string, body interface{}) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	url := baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", notionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("notion API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return json.RawMessage(respBody), nil
}

// CreatePage creates a new page under a parent.
func (c *Client) CreatePage(ctx context.Context, payload map[string]interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, "/pages", payload)
}

// CreateDatabase creates an inline database inside a page.
func (c *Client) CreateDatabase(ctx context.Context, payload map[string]interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, "/databases", payload)
}

// QueryDatabase queries a Notion database with optional filter/sort.
func (c *Client) QueryDatabase(ctx context.Context, dbID string, payload map[string]interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPost, "/databases/"+dbID+"/query", payload)
}

// UpdatePage updates properties of a Notion page.
func (c *Client) UpdatePage(ctx context.Context, pageID string, payload map[string]interface{}) (json.RawMessage, error) {
	return c.do(ctx, http.MethodPatch, "/pages/"+pageID, payload)
}

// AppendBlocks appends content blocks (e.g. images) to a page.
func (c *Client) AppendBlocks(ctx context.Context, pageID string, children []map[string]interface{}) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"children": children,
	}
	return c.do(ctx, http.MethodPatch, "/blocks/"+pageID+"/children", payload)
}

// GetPage retrieves a page by ID.
func (c *Client) GetPage(ctx context.Context, pageID string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/pages/"+pageID, nil)
}

// SearchPages searches for pages with a given title query.
func (c *Client) SearchPages(ctx context.Context, query string) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"query": query,
		"filter": map[string]interface{}{
			"value":    "page",
			"property": "object",
		},
	}
	return c.do(ctx, http.MethodPost, "/search", payload)
}
