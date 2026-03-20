package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// MemberRepo implements ports.MemberRepository using Notion.
type MemberRepo struct {
	client   *Client
	parentID string // parent page ID for the Trade Journal workspace

	// In-memory cache: telegramID → Member (to avoid repeated API calls).
	mu    sync.RWMutex
	cache map[int64]*domain.Member
}

// NewMemberRepo creates a MemberRepo.
func NewMemberRepo(client *Client, parentID string) *MemberRepo {
	return &MemberRepo{
		client:   client,
		parentID: parentID,
		cache:    make(map[int64]*domain.Member),
	}
}

// EnsureMember creates the member page + inline trade database if it doesn't exist.
func (r *MemberRepo) EnsureMember(ctx context.Context, m *domain.Member) (*domain.Member, error) {
	// Check cache first
	r.mu.RLock()
	if cached, ok := r.cache[m.TelegramID]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	// Search Notion for existing page with matching title
	existing, err := r.findMemberPage(ctx, m.TelegramID)
	if err == nil && existing != nil {
		r.mu.Lock()
		r.cache[m.TelegramID] = existing
		r.mu.Unlock()
		return existing, nil
	}

	// Create member page
	title := fmt.Sprintf("@%s — Trade Journal", m.Username)
	if m.Username == "" {
		title = fmt.Sprintf("%s — Trade Journal", m.FirstName)
	}

	pagePayload := map[string]interface{}{
		"parent": map[string]interface{}{
			"page_id": r.parentID,
		},
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]interface{}{"content": title}},
				},
			},
		},
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "callout",
				"callout": map[string]interface{}{
					"icon": map[string]interface{}{"emoji": "📊"},
					"rich_text": []map[string]interface{}{
						{"text": map[string]interface{}{"content": fmt.Sprintf("Telegram ID: %d | Username: @%s", m.TelegramID, m.Username)}},
					},
				},
			},
		},
	}

	raw, err := r.client.CreatePage(ctx, pagePayload)
	if err != nil {
		return nil, fmt.Errorf("create member page: %w", err)
	}

	var pageResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &pageResp); err != nil {
		return nil, fmt.Errorf("parse page response: %w", err)
	}
	m.NotionPageID = pageResp.ID

	// Create inline database for trades inside this member page
	dbID, err := r.createTradeDatabase(ctx, pageResp.ID)
	if err != nil {
		return nil, fmt.Errorf("create trade db: %w", err)
	}
	m.NotionDBID = dbID

	r.mu.Lock()
	r.cache[m.TelegramID] = m
	r.mu.Unlock()

	return m, nil
}

// GetMember retrieves a member by Telegram ID.
func (r *MemberRepo) GetMember(ctx context.Context, telegramID int64) (*domain.Member, error) {
	r.mu.RLock()
	if cached, ok := r.cache[telegramID]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	m, err := r.findMemberPage(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("member %d not found", telegramID)
	}
	r.mu.Lock()
	r.cache[telegramID] = m
	r.mu.Unlock()
	return m, nil
}

// ListMembers returns all cached members. For a full listing, we search Notion.
func (r *MemberRepo) ListMembers(ctx context.Context) ([]domain.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	members := make([]domain.Member, 0, len(r.cache))
	for _, m := range r.cache {
		members = append(members, *m)
	}
	return members, nil
}

// findMemberPage searches for a member page by Telegram ID in the callout text.
func (r *MemberRepo) findMemberPage(ctx context.Context, telegramID int64) (*domain.Member, error) {
	query := fmt.Sprintf("Telegram ID: %d", telegramID)
	raw, err := r.client.SearchPages(ctx, query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}

	if len(result.Results) == 0 {
		return nil, nil
	}

	// Use the first match
	pageID := result.Results[0].ID

	// We need to find the inline database inside this page.
	// For now, store the page ID; the trade DB ID will be resolved lazily.
	return &domain.Member{
		TelegramID:   telegramID,
		NotionPageID: pageID,
	}, nil
}

// createTradeDatabase creates the inline Notion database for trades.
func (r *MemberRepo) createTradeDatabase(ctx context.Context, parentPageID string) (string, error) {
	dbPayload := map[string]interface{}{
		"parent": map[string]interface{}{
			"page_id": parentPageID,
		},
		"title": []map[string]interface{}{
			{"text": map[string]interface{}{"content": "Trades"}},
		},
		"is_inline": true,
		"properties": map[string]interface{}{
			"Symbol": map[string]interface{}{
				"title": map[string]interface{}{},
			},
			"Date": map[string]interface{}{
				"date": map[string]interface{}{},
			},
			"Asset Type": map[string]interface{}{
				"select": map[string]interface{}{
					"options": []map[string]interface{}{
						{"name": "Forex", "color": "blue"},
						{"name": "Gold", "color": "yellow"},
						{"name": "Indices", "color": "green"},
						{"name": "Crypto", "color": "purple"},
					},
				},
			},
			"Direction": map[string]interface{}{
				"select": map[string]interface{}{
					"options": []map[string]interface{}{
						{"name": "BUY", "color": "green"},
						{"name": "SELL", "color": "red"},
					},
				},
			},
			"Status": map[string]interface{}{
				"select": map[string]interface{}{
					"options": []map[string]interface{}{
						{"name": "OPEN", "color": "default"},
						{"name": "WIN", "color": "green"},
						{"name": "LOSS", "color": "red"},
						{"name": "BE", "color": "gray"},
					},
				},
			},
			"Result RR": map[string]interface{}{
				"number": map[string]interface{}{"format": "number"},
			},
			"Time Window": map[string]interface{}{
				"select": map[string]interface{}{
					"options": []map[string]interface{}{
						{"name": "Asia", "color": "blue"},
						{"name": "London", "color": "green"},
						{"name": "NY AM", "color": "yellow"},
						{"name": "NY PM", "color": "orange"},
					},
				},
			},
			"Confluence": map[string]interface{}{
				"rich_text": map[string]interface{}{},
			},
			"Notes": map[string]interface{}{
				"rich_text": map[string]interface{}{},
			},
		},
	}

	raw, err := r.client.CreateDatabase(ctx, dbPayload)
	if err != nil {
		return "", err
	}

	var dbResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &dbResp); err != nil {
		return "", err
	}
	return dbResp.ID, nil
}
