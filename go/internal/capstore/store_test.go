package capstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nkramber/workout-app/go/internal/ai"
)

// Store must stay a cap hook of the role layer.
var _ ai.CapHook = (*Store)(nil)

// TestMonth: the period is the calendar month in UTC (D-190). A time
// late on the last day of a month in New York is in the next month in
// UTC.
func TestMonth(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("no time zone data:", err)
	}
	for in, want := range map[time.Time]string{
		time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC):                      "2026-10",
		time.Date(2026, 9, 30, 23, 59, 59, 999, time.UTC):                 "2026-09",
		time.Date(2026, 9, 30, 21, 0, 0, 0, ny):                           "2026-10",
		time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC):                  "2026-12",
		time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond): "2026-12",
	} {
		if got := Month(in); got != want {
			t.Errorf("Month(%v) = %q, want %q", in, got, want)
		}
	}
}

// TestReserveRefusesBadInput: a bad uid or a negative reservation gives
// an error before the store is read.
func TestReserveRefusesBadInput(t *testing.T) {
	s := New(nil, ai.Caps{User: ai.USD, Project: 2 * ai.USD})
	ctx := context.Background()
	for _, uid := range []string{"", "a/b"} {
		if _, err := s.Reserve(ctx, uid, 1); !errors.Is(err, errUID) {
			t.Errorf("Reserve(%q) = %v, want errUID", uid, err)
		}
	}
	if _, err := s.Reserve(ctx, "uid-1", -1); !errors.Is(err, errWorst) {
		t.Errorf("Reserve(-1) = %v, want errWorst", err)
	}
}
