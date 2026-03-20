package notion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// TradeRepo implements ports.TradeRepository using Notion.
type TradeRepo struct {
	client    *Client
	memberRepo *MemberRepo
}

// NewTradeRepo creates a TradeRepo.
func NewTradeRepo(client *Client, memberRepo *MemberRepo) *TradeRepo {
	return &TradeRepo{client: client, memberRepo: memberRepo}
}

// SaveTrade creates a new trade entry in the member's inline database.
func (r *TradeRepo) SaveTrade(ctx context.Context, memberID int64, trade *domain.Trade) error {
	member, err := r.memberRepo.GetMember(ctx, memberID)
	if err != nil {
		return fmt.Errorf("get member: %w", err)
	}
	if member.NotionDBID == "" {
		return fmt.Errorf("member %d has no trade database", memberID)
	}

	props := r.buildTradeProperties(trade)
	payload := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": member.NotionDBID,
		},
		"properties": props,
	}

	raw, err := r.client.CreatePage(ctx, payload)
	if err != nil {
		return fmt.Errorf("create trade page: %w", err)
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	trade.ID = resp.ID
	return nil
}

// GetTrades retrieves all trades for a member from their inline database.
func (r *TradeRepo) GetTrades(ctx context.Context, memberID int64) ([]domain.Trade, error) {
	member, err := r.memberRepo.GetMember(ctx, memberID)
	if err != nil {
		return nil, fmt.Errorf("get member: %w", err)
	}
	if member.NotionDBID == "" {
		return nil, nil
	}

	payload := map[string]interface{}{
		"sorts": []map[string]interface{}{
			{"property": "Date", "direction": "descending"},
		},
	}

	raw, err := r.client.QueryDatabase(ctx, member.NotionDBID, payload)
	if err != nil {
		return nil, fmt.Errorf("query trades: %w", err)
	}

	return r.parseTrades(raw)
}

// GetTradeByID fetches a single trade page from Notion.
func (r *TradeRepo) GetTradeByID(ctx context.Context, tradeID string) (*domain.Trade, error) {
	raw, err := r.client.GetPage(ctx, tradeID)
	if err != nil {
		return nil, fmt.Errorf("get trade page: %w", err)
	}
	return r.parseSingleTrade(raw)
}

// UpdateTrade patches a trade's properties.
func (r *TradeRepo) UpdateTrade(ctx context.Context, trade *domain.Trade) error {
	props := r.buildTradeProperties(trade)
	payload := map[string]interface{}{
		"properties": props,
	}
	_, err := r.client.UpdatePage(ctx, trade.ID, payload)
	return err
}

// buildTradeProperties converts a Trade to Notion page properties.
func (r *TradeRepo) buildTradeProperties(t *domain.Trade) map[string]interface{} {
	props := map[string]interface{}{
		"Symbol": map[string]interface{}{
			"title": []map[string]interface{}{
				{"text": map[string]interface{}{"content": t.Symbol}},
			},
		},
		"Direction": map[string]interface{}{
			"select": map[string]interface{}{"name": string(t.Direction)},
		},
		"Asset Type": map[string]interface{}{
			"select": map[string]interface{}{"name": t.AssetType.String()},
		},
		"Entry Price": map[string]interface{}{
			"number": t.EntryPrice,
		},
		"Stop Loss": map[string]interface{}{
			"number": t.StopLoss,
		},
		"Take Profit": map[string]interface{}{
			"number": t.TakeProfit,
		},
		"Status": map[string]interface{}{
			"select": map[string]interface{}{"name": string(t.Status)},
		},
	}

	if !t.Date.IsZero() {
		props["Date"] = map[string]interface{}{
			"date": map[string]interface{}{
				"start": t.Date.Format("2006-01-02"),
			},
		}
	}
	if t.ClosePrice != 0 {
		props["Close Price"] = map[string]interface{}{"number": t.ClosePrice}
	}
	if t.ResultPips != 0 {
		props["Result Pips"] = map[string]interface{}{"number": t.ResultPips}
	}
	if t.RRRatio != 0 {
		props["RR Ratio"] = map[string]interface{}{"number": t.RRRatio}
	}
	if t.Notes != "" {
		props["Notes"] = map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"text": map[string]interface{}{"content": t.Notes}},
			},
		}
	}
	return props
}

// parseTrades extracts trades from a Notion database query response.
func (r *TradeRepo) parseTrades(raw json.RawMessage) ([]domain.Trade, error) {
	var resp struct {
		Results []json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}

	trades := make([]domain.Trade, 0, len(resp.Results))
	for _, item := range resp.Results {
		t, err := r.parseSingleTrade(item)
		if err != nil {
			continue // skip malformed entries
		}
		trades = append(trades, *t)
	}
	return trades, nil
}

// parseSingleTrade extracts a Trade from a Notion page JSON.
func (r *TradeRepo) parseSingleTrade(raw json.RawMessage) (*domain.Trade, error) {
	var page struct {
		ID         string `json:"id"`
		Properties struct {
			Symbol struct {
				Title []struct {
					PlainText string `json:"plain_text"`
				} `json:"title"`
			} `json:"Symbol"`
			Date struct {
				Date *struct {
					Start string `json:"start"`
				} `json:"date"`
			} `json:"Date"`
			AssetType struct {
				Select *struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Asset Type"`
			Direction struct {
				Select *struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Direction"`
			EntryPrice struct {
				Number *float64 `json:"number"`
			} `json:"Entry Price"`
			StopLoss struct {
				Number *float64 `json:"number"`
			} `json:"Stop Loss"`
			TakeProfit struct {
				Number *float64 `json:"number"`
			} `json:"Take Profit"`
			ClosePrice struct {
				Number *float64 `json:"number"`
			} `json:"Close Price"`
			ResultPips struct {
				Number *float64 `json:"number"`
			} `json:"Result Pips"`
			RRRatio struct {
				Number *float64 `json:"number"`
			} `json:"RR Ratio"`
			Status struct {
				Select *struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Status"`
			Notes struct {
				RichText []struct {
					PlainText string `json:"plain_text"`
				} `json:"rich_text"`
			} `json:"Notes"`
		} `json:"properties"`
	}

	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}

	t := &domain.Trade{ID: page.ID}

	if len(page.Properties.Symbol.Title) > 0 {
		t.Symbol = page.Properties.Symbol.Title[0].PlainText
	}
	if page.Properties.Date.Date != nil {
		// best-effort parse
		if parsed, err := parseDate(page.Properties.Date.Date.Start); err == nil {
			t.Date = parsed
		}
	}
	if page.Properties.AssetType.Select != nil {
		t.AssetType = domain.AssetType(page.Properties.AssetType.Select.Name)
	}
	if page.Properties.Direction.Select != nil {
		t.Direction = domain.Direction(page.Properties.Direction.Select.Name)
	}
	if page.Properties.EntryPrice.Number != nil {
		t.EntryPrice = *page.Properties.EntryPrice.Number
	}
	if page.Properties.StopLoss.Number != nil {
		t.StopLoss = *page.Properties.StopLoss.Number
	}
	if page.Properties.TakeProfit.Number != nil {
		t.TakeProfit = *page.Properties.TakeProfit.Number
	}
	if page.Properties.ClosePrice.Number != nil {
		t.ClosePrice = *page.Properties.ClosePrice.Number
	}
	if page.Properties.ResultPips.Number != nil {
		t.ResultPips = *page.Properties.ResultPips.Number
	}
	if page.Properties.RRRatio.Number != nil {
		t.RRRatio = *page.Properties.RRRatio.Number
	}
	if page.Properties.Status.Select != nil {
		t.Status = domain.TradeStatus(page.Properties.Status.Select.Name)
	}
	if len(page.Properties.Notes.RichText) > 0 {
		t.Notes = page.Properties.Notes.RichText[0].PlainText
	}

	return t, nil
}
