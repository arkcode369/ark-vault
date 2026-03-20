package badger

import (
	"context"
	"encoding/json"
	"fmt"

	badgerdb "github.com/dgraph-io/badger/v4"

	"github.com/arkcode369/ark-vault/internal/domain"
)

// ReminderRepo implements ports.ReminderStore backed by BadgerDB.
type ReminderRepo struct {
	store *Store
}

// NewReminderRepo creates a new ReminderRepo.
func NewReminderRepo(store *Store) *ReminderRepo {
	return &ReminderRepo{store: store}
}

func reminderPrefKey(telegramID int64) string {
	return fmt.Sprintf("gam:reminder_pref:%d", telegramID)
}

const reminderPrefPrefix = "gam:reminder_pref:"

// GetReminderPref returns the reminder preference for the given Telegram ID.
// Returns nil (not an error) when no preference exists yet.
func (r *ReminderRepo) GetReminderPref(_ context.Context, telegramID int64) (*domain.ReminderPreference, error) {
	var p domain.ReminderPreference
	err := r.store.Get(reminderPrefKey(telegramID), &p)
	if err == badgerdb.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SaveReminderPref persists a reminder preference.
func (r *ReminderRepo) SaveReminderPref(_ context.Context, pref *domain.ReminderPreference) error {
	return r.store.Set(reminderPrefKey(pref.TelegramID), pref)
}

// ListEnabledReminders returns all reminder preferences where Enabled is true.
func (r *ReminderRepo) ListEnabledReminders(_ context.Context) ([]domain.ReminderPreference, error) {
	raw, err := r.store.Scan(reminderPrefPrefix)
	if err != nil {
		return nil, err
	}
	var prefs []domain.ReminderPreference
	for _, b := range raw {
		var p domain.ReminderPreference
		if err := json.Unmarshal(b, &p); err != nil {
			return nil, err
		}
		if p.Enabled {
			prefs = append(prefs, p)
		}
	}
	return prefs, nil
}
