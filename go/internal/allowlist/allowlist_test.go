package allowlist

import (
	"context"
	"errors"
	"testing"
	"time"
)

type counter struct {
	calls int
	uids  map[string]bool
	err   error
}

func (c *counter) lookup(_ context.Context, uid string) (bool, error) {
	c.calls++
	return c.uids[uid], c.err
}

func TestAllowed(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	src := &counter{uids: map[string]bool{"uid-a": true}}
	list := New(src.lookup).WithClock(func() time.Time { return now })

	if ok, err := list.Allowed(ctx, "uid-a"); !ok || err != nil {
		t.Fatalf("Allowed(uid-a) = %v, %v, want true", ok, err)
	}
	if ok, err := list.Allowed(ctx, "uid-b"); ok || err != nil {
		t.Fatalf("Allowed(uid-b) = %v, %v, want false", ok, err)
	}
	if ok, _ := list.Allowed(ctx, ""); ok {
		t.Fatal("Allowed(\"\") = true, want false")
	}
	if src.calls != 2 {
		t.Fatalf("lookups = %d, want 2: an empty uid needs no read", src.calls)
	}

	// Inside the TTL the cache answers, also for a changed list.
	src.uids = map[string]bool{"uid-b": true}
	now = now.Add(TTL - time.Second)
	if ok, _ := list.Allowed(ctx, "uid-a"); !ok {
		t.Fatal("Allowed(uid-a) inside the TTL = false, want the cached true")
	}
	if src.calls != 2 {
		t.Fatalf("lookups = %d, want 2 inside the TTL", src.calls)
	}

	// After the TTL the list reads again, so a change takes effect.
	now = now.Add(time.Second)
	if ok, _ := list.Allowed(ctx, "uid-a"); ok {
		t.Fatal("Allowed(uid-a) after the TTL = true, want false")
	}
	if ok, _ := list.Allowed(ctx, "uid-b"); !ok {
		t.Fatal("Allowed(uid-b) after the TTL = false, want true")
	}
}

func TestAllowedErrorIsNotCached(t *testing.T) {
	ctx := context.Background()
	src := &counter{uids: map[string]bool{"uid-a": true}, err: errors.New("down")}
	list := New(src.lookup)
	if ok, err := list.Allowed(ctx, "uid-a"); ok || err == nil {
		t.Fatalf("Allowed = %v, %v, want false and the error", ok, err)
	}
	src.err = nil
	if ok, err := list.Allowed(ctx, "uid-a"); !ok || err != nil {
		t.Fatalf("Allowed after recovery = %v, %v, want true", ok, err)
	}
	if src.calls != 2 {
		t.Fatalf("lookups = %d, want 2", src.calls)
	}
}
