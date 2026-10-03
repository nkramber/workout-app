package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestFakeDelay: a fake with a delay waits before its reply, and the end
// of the context stops the wait (D-241).
func TestFakeDelay(t *testing.T) {
	f := &Fake{Delay: 50 * time.Millisecond, Reply: func(Call) (Reply, error) { return Reply{Text: "ok"}, nil }}
	start := time.Now()
	r, err := f.Send(context.Background(), Call{})
	if err != nil || r.Text != "ok" || time.Since(start) < f.Delay {
		t.Fatalf("reply %+v, err %v, after %v", r, err, time.Since(start))
	}

	f.Delay = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.Send(ctx, Call{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v, want context.Canceled", err)
	}
}
