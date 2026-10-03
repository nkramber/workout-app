// Package capstore holds the lasting cap hook of the Luna role layer
// (D-189). Firestore holds the spend of each period, so a new instance
// of the API reads the spend of the earlier instances. The period is
// the calendar month in UTC (D-190), and the caps come from the
// configuration (D-188).
package capstore

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/ai"
)

// The Firestore paths of the spend of one month are
// users/{uid}/aiSpend/{YYYY-MM} for the user and aiSpend/{YYYY-MM} for
// the project (D-224). One transaction reads and writes both.
const (
	UsersCollection = "users"
	Collection      = "aiSpend"
	MonthLayout     = "2006-01"
)

// settleTimeout limits the transaction of a settle. A settle runs after
// the call, also when the request of the user ends first, so that the
// charge reaches the store.
const settleTimeout = 10 * time.Second

// maxAttempts is the number of attempts of one transaction. Two calls
// at the same time lock the same documents, and Firestore aborts one
// transaction, which the client then tries again. When each attempt
// aborts, the error stops the call, so it never passes the cap.
const maxAttempts = 20

// Month gives the period of t: the calendar month in UTC, such as
// "2026-10" (D-190).
func Month(t time.Time) string { return t.UTC().Format(MonthLayout) }

var (
	errUID   = errors.New("capstore: a uid of 1 or more characters with no slash is required")
	errWorst = errors.New("capstore: a reservation of 0 or more is required")
)

// Store is the cap hook of the API. Reserve adds the worst-case cost of
// a call to the reservations of the user and of the project in one
// transaction, or gives ai.ErrCap when a cap can not cover it. Settle
// moves the reservation to the charge in a second transaction. When the
// API stops between the two, the reservation stays, so the spend can be
// too high but never too low. It is safe for use by more than one
// goroutine and by more than one instance of the API.
type Store struct {
	client *firestore.Client
	caps   ai.Caps
	now    func() time.Time
}

// New gives the cap hook of the client with the caps of each month.
func New(client *firestore.Client, caps ai.Caps) *Store {
	return &Store{client: client, caps: caps, now: time.Now}
}

// The stored document of the spend of one month. Charged is the sum of
// the settled costs. Reserved is the sum of the open reservations.
type spend struct {
	Charged  int64 `firestore:"charged_nano_usd"`
	Reserved int64 `firestore:"reserved_nano_usd"`
}

func (s spend) total() ai.NanoUSD { return ai.NanoUSD(s.Charged + s.Reserved) }

func (s *Store) refs(uid, month string) (user, project *firestore.DocumentRef) {
	user = s.client.Collection(UsersCollection).Doc(uid).Collection(Collection).Doc(month)
	project = s.client.Collection(Collection).Doc(month)
	return user, project
}

// read gives the spend of each document. A document that does not
// exist gives a spend of 0, so the spend of a new month starts at 0.
func read(tx *firestore.Transaction, refs ...*firestore.DocumentRef) ([]spend, error) {
	snaps, err := tx.GetAll(refs)
	if err != nil {
		return nil, err
	}
	out := make([]spend, len(snaps))
	for i, snap := range snaps {
		if !snap.Exists() {
			continue
		}
		if err := snap.DataTo(&out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// update reads the spend of the user and of the project in one
// transaction, applies change to each, and writes both.
func (s *Store) update(ctx context.Context, uid, month string, change func(user, project *spend) error) error {
	user, project := s.refs(uid, month)
	return s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		got, err := read(tx, user, project)
		if err != nil {
			return err
		}
		if err := change(&got[0], &got[1]); err != nil {
			return err
		}
		if err := tx.Set(user, got[0]); err != nil {
			return err
		}
		return tx.Set(project, got[1])
	}, firestore.MaxAttempts(maxAttempts))
}

// Reserve reserves the worst-case cost of a call in the current month,
// or gives ai.ErrCap when the spend and the reservations of the user or
// of the project can not cover it. Another error tells that the store
// failed, and the client then makes no call.
func (s *Store) Reserve(ctx context.Context, uid string, worst ai.NanoUSD) (func(ai.NanoUSD) error, error) {
	if uid == "" || strings.Contains(uid, "/") {
		return nil, errUID
	}
	if worst < 0 {
		return nil, errWorst
	}
	month := Month(s.now())
	err := s.update(ctx, uid, month, func(u, p *spend) error {
		if u.total()+worst > s.caps.User || p.total()+worst > s.caps.Project {
			return ai.ErrCap
		}
		u.Reserved += int64(worst)
		p.Reserved += int64(worst)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// The settle applies to the month of the reservation, also after the
	// end of that month.
	base := context.WithoutCancel(ctx)
	var once sync.Once
	var settleErr error
	return func(cost ai.NanoUSD) error {
		once.Do(func() {
			ctx, cancel := context.WithTimeout(base, settleTimeout)
			defer cancel()
			settleErr = s.update(ctx, uid, month, func(u, p *spend) error {
				u.Reserved -= int64(worst)
				p.Reserved -= int64(worst)
				u.Charged += int64(cost)
				p.Charged += int64(cost)
				return nil
			})
		})
		return settleErr
	}, nil
}
