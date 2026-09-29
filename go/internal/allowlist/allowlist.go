// Package allowlist holds the uids that can use the API (D-75, D-131).
// Firestore holds one document for each allowed uid, in the collection
// allowlist, with the uid as the document id. The document can be empty.
// No uid goes into the repository: the owner writes the entry in the
// live project. A cache of one minute keeps a call from a Firestore read,
// so a change of the list takes effect in one minute or less.
package allowlist

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Collection names the Firestore collection of the list.
const Collection = "allowlist"

// TTL is the age at which a cached answer expires.
const TTL = time.Minute

// Lookup reads whether one uid is on the list. Firestore is the one
// production source, and tests give a function.
type Lookup func(ctx context.Context, uid string) (bool, error)

type entry struct {
	allowed bool
	expires time.Time
}

// List answers Allowed from a cached lookup.
type List struct {
	lookup Lookup
	now    func() time.Time

	mu    sync.Mutex
	cache map[string]entry
}

// New builds a list over a lookup.
func New(lookup Lookup) *List {
	return &List{lookup: lookup, now: time.Now, cache: map[string]entry{}}
}

// WithClock replaces time.Now (tests).
func (l *List) WithClock(now func() time.Time) *List {
	l.now = now
	return l
}

// Allowed reports whether the uid is on the list. An empty uid is never
// on it. A lookup error is not cached, so the next call reads again.
func (l *List) Allowed(ctx context.Context, uid string) (bool, error) {
	if uid == "" {
		return false, nil
	}
	now := l.now()
	l.mu.Lock()
	e, ok := l.cache[uid]
	l.mu.Unlock()
	if ok && now.Before(e.expires) {
		return e.allowed, nil
	}
	allowed, err := l.lookup(ctx, uid)
	if err != nil {
		return false, err
	}
	l.mu.Lock()
	l.cache[uid] = entry{allowed: allowed, expires: now.Add(TTL)}
	l.mu.Unlock()
	return allowed, nil
}

// FromFirestore builds the production list. A uid is on the list when
// its document exists.
func FromFirestore(client *firestore.Client) *List {
	return New(func(ctx context.Context, uid string) (bool, error) {
		_, err := client.Collection(Collection).Doc(uid).Get(ctx)
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	})
}
