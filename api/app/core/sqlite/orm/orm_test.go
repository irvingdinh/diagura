package orm

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGenerateUUIDv7Format(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, id string)
	}{
		{
			name: "length is 36",
			check: func(t *testing.T, id string) {
				if len(id) != 36 {
					t.Errorf("len = %d, want 36", len(id))
				}
			},
		},
		{
			name: "dashes at correct positions",
			check: func(t *testing.T, id string) {
				for _, pos := range []int{8, 13, 18, 23} {
					if id[pos] != '-' {
						t.Errorf("id[%d] = %q, want '-'", pos, id[pos])
					}
				}
			},
		},
		{
			name: "version nibble is 7",
			check: func(t *testing.T, id string) {
				if id[14] != '7' {
					t.Errorf("version char = %q, want '7'", id[14])
				}
			},
		},
		{
			name: "variant bits are 10xx",
			check: func(t *testing.T, id string) {
				c := id[19]
				if c != '8' && c != '9' && c != 'a' && c != 'b' {
					t.Errorf("variant char = %q, want one of 8/9/a/b", c)
				}
			},
		},
		{
			name: "all non-dash chars are hex",
			check: func(t *testing.T, id string) {
				for i, c := range id {
					if i == 8 || i == 13 || i == 18 || i == 23 {
						continue
					}
					if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
						t.Errorf("id[%d] = %q, not a hex char", i, c)
					}
				}
			},
		},
	}

	id := NewID()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, id)
		})
	}
}

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

func TestGenerateUUIDv7ConcurrentUniqueness(t *testing.T) {
	const goroutines = 8
	const idsPerGoroutine = 100_000

	var mu sync.Mutex
	seen := make(map[string]struct{}, goroutines*idsPerGoroutine)
	var dupCount int

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()

			local := make([]string, idsPerGoroutine)
			for i := range local {
				local[i] = NewID()
			}

			// Verify per-goroutine monotonicity.
			for i := 1; i < len(local); i++ {
				if local[i] <= local[i-1] {
					t.Errorf("per-goroutine monotonicity broken at %d: %q <= %q", i, local[i], local[i-1])
					return
				}
			}

			mu.Lock()
			for _, id := range local {
				if _, ok := seen[id]; ok {
					dupCount++
				}
				seen[id] = struct{}{}
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if dupCount > 0 {
		t.Errorf("found %d duplicate IDs out of %d total", dupCount, goroutines*idsPerGoroutine)
	}
}

func BenchmarkNewID(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		var id string
		for b.Loop() {
			id = NewID()
		}
		runtime.KeepAlive(id)
	})

	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var id string
			for pb.Next() {
				id = NewID()
			}
			runtime.KeepAlive(id)
		})
	})
}
