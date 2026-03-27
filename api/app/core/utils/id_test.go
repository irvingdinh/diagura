package utils

import (
	"testing"
	"time"
)

func TestGenerateUUIDv7Monotonicity(t *testing.T) {
	now := time.Now()

	prev := generateUUIDv7(now)
	for i := 0; i < 1000; i++ {
		next := generateUUIDv7(now)
		if next <= prev {
			t.Fatalf("iteration %d: %q <= %q (not monotonic)", i, next, prev)
		}
		prev = next
	}
}

func TestGenerateUUIDv7ClockRegression(t *testing.T) {
	future := time.Now().Add(10 * time.Second)
	past := time.Now().Add(-10 * time.Second)

	futureID := generateUUIDv7(future)
	pastID := generateUUIDv7(past)

	if pastID <= futureID {
		// Clock regression should still produce a greater ID because
		// the counter advances and lastMS is reused.
		t.Logf("futureID=%s pastID=%s", futureID, pastID)
	}
	// The key invariant: pastID must be > futureID because the
	// monotonic counter ensures forward progress.
	if pastID <= futureID {
		t.Errorf("clock regression broke monotonicity: past=%q <= future=%q", pastID, futureID)
	}
}

func TestGenerateUUIDv7CounterOverflow(t *testing.T) {
	// Reset state for a clean test.
	uuidState.Lock()
	uuidState.lastMS = 0
	uuidState.counter = 0
	uuidState.Unlock()

	now := time.Now()

	// Generate enough IDs to overflow the 12-bit counter (max 0x0FFF = 4095).
	// The initial seed uses mask 0x01FF (max 511), so we need at most
	// 4096 - 0 = 4096 IDs to overflow.
	ids := make([]string, 4500)
	for i := range ids {
		ids[i] = generateUUIDv7(now)
	}

	// Verify all are monotonically increasing (overflow handled correctly).
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			t.Fatalf("monotonicity broken at index %d: %q <= %q", i, ids[i], ids[i-1])
		}
	}

	// Verify all are unique.
	seen := make(map[string]struct{}, len(ids))
	for i, id := range ids {
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate at index %d: %q", i, id)
		}
		seen[id] = struct{}{}
	}
}
