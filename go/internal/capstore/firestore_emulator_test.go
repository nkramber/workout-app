//go:build emulator

// The lasting cap hook over the emulator. `make emulator-test` starts
// the emulators and runs this file with the tag emulator.
package capstore

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/ai"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// The caps of D-188.
var monthly = ai.Caps{User: ai.USD, Project: 2 * ai.USD}

// newClient gives a new Firestore client. A new client and a new Store
// stand in for a new instance of the API.
func newClient(t *testing.T) *firestore.Client {
	t.Helper()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" || os.Getenv("GOOGLE_CLOUD_PROJECT") != "demo-workout-app" {
		t.Fatal("run this test with `make emulator-test`, which sets FIRESTORE_EMULATOR_HOST and the demo project")
	}
	client, err := firestore.NewClient(context.Background(), "demo-workout-app")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// clock is a time that a test sets.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time { c.mu.Lock(); defer c.mu.Unlock(); return c.t }

func (c *clock) set(t time.Time) { c.mu.Lock(); defer c.mu.Unlock(); c.t = t }

// freshClock gives a clock in a month that no other test uses, because
// all tests share the project document of each month.
func freshClock() *clock {
	return &clock{t: time.Date(2100+rand.IntN(7000), time.March, 15, 12, 0, 0, 0, time.UTC)}
}

func newStore(client *firestore.Client, c *clock) *Store {
	s := New(client, monthly)
	s.now = c.now
	return s
}

func newUID(name string) string { return fmt.Sprintf("cap-%s-%d", name, time.Now().UnixNano()) }

// stored reads the spend document at the path, and fails on a field
// that is not a field of the spend.
func stored(t *testing.T, ref *firestore.DocumentRef) spend {
	t.Helper()
	snap, err := ref.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var keys []string
	for k := range snap.Data() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if fmt.Sprint(keys) != "[charged_nano_usd reserved_nano_usd]" {
		t.Fatalf("stored fields = %v, want charged_nano_usd and reserved_nano_usd", keys)
	}
	var s spend
	if err := snap.DataTo(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

func mustReserve(t *testing.T, s *Store, uid string, worst ai.NanoUSD) func(ai.NanoUSD) error {
	t.Helper()
	settle, err := s.Reserve(context.Background(), uid, worst)
	if err != nil {
		t.Fatalf("Reserve(%s) = %v", worst, err)
	}
	return settle
}

// planRequest gives a planner request of one new dumbbell exercise.
func planRequest(t *testing.T, uid string) ai.Request {
	t.Helper()
	e, ok := domain.DefaultCatalog().Exercise("db_goblet_squat")
	if !ok {
		t.Fatal("db_goblet_squat: not in the catalog")
	}
	set := &domain.DumbbellSet{Lightest: domain.Pounds(5), Heaviest: domain.Pounds(50), Step: domain.Pounds(5)}
	in := policy.Input{Exercise: e, Entry: domain.InventoryEntry{Machine: e.Machine, Dumbbells: set}, Today: "2026-10-01"}
	return ai.Request{User: uid, Today: "2026-10-01", Sessions: 1, Exercises: []policy.Input{in}}
}

// TestRestartRefuses is the acceptance story of the store: after a
// restart of the API, a call over the cap gets ai.ErrCap, and the
// client sends nothing to Luna (D-189).
func TestRestartRefuses(t *testing.T) {
	c := freshClock()
	uid := newUID("restart")
	ctx := context.Background()

	// The first instance charges 0.6 USD and keeps a reservation of
	// 0.3 USD open. Then it stops with no settle.
	first := newClient(t)
	s1 := newStore(first, c)
	if err := mustReserve(t, s1, uid, 7*ai.USD/10)(6 * ai.USD / 10); err != nil {
		t.Fatal(err)
	}
	mustReserve(t, s1, uid, 3*ai.USD/10)
	_ = first.Close()

	// The second instance reads the spend of the month.
	second := newClient(t)
	s2 := newStore(second, c)
	if _, err := s2.Reserve(ctx, uid, 2*ai.USD/10); !errors.Is(err, ai.ErrCap) {
		t.Fatalf("Reserve of 0.2 USD over 0.9 USD of 1 USD = %v, want ErrCap", err)
	}
	settle := mustReserve(t, s2, uid, ai.USD/10)

	user, project := s2.refs(uid, Month(c.now()))
	if got, want := stored(t, user), (spend{Charged: int64(6 * ai.USD / 10), Reserved: int64(4 * ai.USD / 10)}); got != want {
		t.Fatalf("user spend = %+v, want %+v", got, want)
	}
	if got := stored(t, project); got.total() != ai.USD {
		t.Fatalf("project spend = %+v, want 1 USD in total", got)
	}
	if err := settle(ai.USD / 10); err != nil {
		t.Fatal(err)
	}
	if err := settle(0); err != nil { // a second settle changes nothing
		t.Fatal(err)
	}
	if got := stored(t, user); got.Charged != int64(7*ai.USD/10) || got.total() != ai.USD {
		t.Fatalf("user spend after the settle = %+v, want 0.7 USD charged and 1 USD in total", got)
	}

	// Through the role layer, a planner call over the cap sends nothing.
	fake := &ai.Fake{}
	var recs []ai.CostRecord
	client := &ai.Client{Provider: fake, Cap: s2, Record: func(r ai.CostRecord) { recs = append(recs, r) }}
	res, err := client.Plan(ctx, planRequest(t, uid))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != ai.StatusCapped || len(fake.Calls()) != 0 || len(recs) != 1 || recs[0].Reserved != 0 {
		t.Fatalf("status %q with %d calls and records %+v: want %q with no call", res.Status, len(fake.Calls()), recs, ai.StatusCapped)
	}
}

// TestNewMonth: the spend of a new month starts at 0. A settle after
// the end of a month charges the month of its reservation.
func TestNewMonth(t *testing.T) {
	c := freshClock()
	uid := newUID("month")
	s := newStore(newClient(t), c)
	ctx := context.Background()
	march := Month(c.now())

	if err := mustReserve(t, s, uid, ai.USD/2)(ai.USD / 2); err != nil {
		t.Fatal(err)
	}
	late := mustReserve(t, s, uid, ai.USD/2)
	if _, err := s.Reserve(ctx, uid, 1); !errors.Is(err, ai.ErrCap) {
		t.Fatalf("Reserve at the cap = %v, want ErrCap", err)
	}

	// 00:00 UTC on the first day of the next month.
	y, m, _ := c.now().Date()
	c.set(time.Date(y, m+1, 1, 0, 0, 0, 0, time.UTC))
	april := Month(c.now())
	if april == march {
		t.Fatalf("the month did not change: %s", april)
	}
	if err := late(ai.USD / 4); err != nil {
		t.Fatal(err)
	}
	if err := mustReserve(t, s, uid, ai.USD)(ai.USD / 10); err != nil {
		t.Fatal(err)
	}

	userMarch, _ := s.refs(uid, march)
	userApril, projectApril := s.refs(uid, april)
	if got, want := stored(t, userMarch), (spend{Charged: int64(3 * ai.USD / 4)}); got != want {
		t.Fatalf("spend of %s = %+v, want %+v", march, got, want)
	}
	if got, want := stored(t, userApril), (spend{Charged: int64(ai.USD / 10)}); got != want {
		t.Fatalf("spend of %s = %+v, want %+v", april, got, want)
	}
	if got, want := stored(t, projectApril), (spend{Charged: int64(ai.USD / 10)}); got != want {
		t.Fatalf("project spend of %s = %+v, want %+v", april, got, want)
	}
}

// TestParallel: calls at the same time, from two instances of the API,
// never pass a cap together. Each reservation of 0.3 USD fits three
// times in the user cap of 1 USD, and six times in the project cap of
// 2 USD.
func TestParallel(t *testing.T) {
	for _, tc := range []struct {
		name string
		uids int
		want int
	}{
		{"user", 1, 3},
		{"project", 10, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := freshClock()
			instances := []*Store{newStore(newClient(t), c), newStore(newClient(t), c)}
			worst := 3 * ai.USD / 10
			var wg sync.WaitGroup
			var mu sync.Mutex
			passed := 0
			var others []error
			for i := range 10 {
				uid := fmt.Sprintf("cap-parallel-%d-%d", c.now().Year(), i%tc.uids)
				wg.Go(func() {
					_, err := instances[i%2].Reserve(context.Background(), uid, worst)
					mu.Lock()
					defer mu.Unlock()
					switch {
					case err == nil:
						passed++
					case !errors.Is(err, ai.ErrCap):
						others = append(others, err)
					}
				})
			}
			wg.Wait()
			if passed != tc.want || len(others) != 0 {
				t.Fatalf("%d reservations passed with the errors %v: want %d and no other error", passed, others, tc.want)
			}
			_, project := instances[0].refs("any", Month(c.now()))
			if got := stored(t, project); got.Reserved != int64(tc.want)*int64(worst) || got.Charged != 0 {
				t.Fatalf("project spend = %+v, want %d reservations", got, tc.want)
			}
		})
	}
}
