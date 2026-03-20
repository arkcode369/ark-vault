package telegram

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Second)
	userID := int64(12345)

	// First 3 should pass
	for i := 0; i < 3; i++ {
		if !rl.Allow(userID) {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 4th should be rejected
	if rl.Allow(userID) {
		t.Error("4th request should be rate limited")
	}

	// Different user should still be allowed
	if !rl.Allow(99999) {
		t.Error("different user should not be rate limited")
	}
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	userID := int64(12345)

	rl.Allow(userID)
	rl.Allow(userID)

	if rl.Allow(userID) {
		t.Error("should be rate limited")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	if !rl.Allow(userID) {
		t.Error("should be allowed after window expiry")
	}
}

func TestRateLimiter_Remaining(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Second)
	userID := int64(12345)

	if rem := rl.Remaining(userID); rem != 5 {
		t.Errorf("remaining: got %d, want 5", rem)
	}

	rl.Allow(userID)
	rl.Allow(userID)

	if rem := rl.Remaining(userID); rem != 3 {
		t.Errorf("remaining: got %d, want 3", rem)
	}
}
