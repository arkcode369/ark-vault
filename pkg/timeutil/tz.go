package timeutil

import "time"

// LoadLocation loads a timezone, falling back to UTC.
func LoadLocation(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// StartOfWeek returns Monday 00:00 of the current week in the given location.
func StartOfWeek(now time.Time, loc *time.Location) time.Time {
	t := now.In(loc)
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday → 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, loc)
}
