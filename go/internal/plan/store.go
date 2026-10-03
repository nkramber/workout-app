package plan

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The Firestore paths of the plan and the exclusions of a user are
// users/{uid}/plan/active and users/{uid}/exclusions/active (D-226).
// Each one is one document.
const (
	UsersCollection      = "users"
	PlanCollection       = "plan"
	ExclusionsCollection = "exclusions"
	ActiveDoc            = "active"
)

// Store reads and saves the plan and the exclusions of a user. Get
// gives false when the user has no plan. Save writes the plan and, when
// replace is true, the list of ex, in one transaction (D-227, D-234). It
// gives ErrConflict, and writes nothing, when the stored exclusions do
// not have the revision of ex. A new list gets the next revision.
type Store interface {
	Get(ctx context.Context, uid string) (Plan, bool, error)
	Exclusions(ctx context.Context, uid string) (Exclusions, error)
	Save(ctx context.Context, uid string, p Plan, ex Exclusions, replace bool) error
}

var errUID = errors.New("plan: a uid of 1 or more characters with no slash is required")

func checkUID(uid string) error {
	if uid == "" || strings.Contains(uid, "/") {
		return errUID
	}
	return nil
}

// Memory is a Store in memory, for tests.
type Memory struct {
	mu         sync.Mutex
	plans      map[string]Plan
	exclusions map[string]Exclusions
}

// NewMemory gives an empty Memory store.
func NewMemory() *Memory {
	return &Memory{plans: map[string]Plan{}, exclusions: map[string]Exclusions{}}
}

// Get gives the plan of the uid.
func (s *Memory) Get(_ context.Context, uid string) (Plan, bool, error) {
	if err := checkUID(uid); err != nil {
		return Plan{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plans[uid]
	return p.clone(), ok, nil
}

// Exclusions gives the exclusions of the uid.
func (s *Memory) Exclusions(_ context.Context, uid string) (Exclusions, error) {
	if err := checkUID(uid); err != nil {
		return Exclusions{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exclusions[uid].clone(), nil
}

// Save stores a copy of the plan, and of the list when replace is true.
func (s *Memory) Save(_ context.Context, uid string, p Plan, ex Exclusions, replace bool) error {
	if err := checkUID(uid); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := s.exclusions[uid]
	if cur.Revision != ex.Revision {
		return ErrConflict
	}
	if replace {
		next := ex.clone()
		next.Revision = cur.Revision + 1
		s.exclusions[uid] = next
	}
	s.plans[uid] = p.clone()
	return nil
}

// Firestore is the production Store.
type Firestore struct {
	client *firestore.Client
}

// FromFirestore gives the Store of the client.
func FromFirestore(client *firestore.Client) *Firestore { return &Firestore{client: client} }

func (s *Firestore) user(uid string) *firestore.DocumentRef {
	return s.client.Collection(UsersCollection).Doc(uid)
}

func (s *Firestore) planRef(uid string) *firestore.DocumentRef {
	return s.user(uid).Collection(PlanCollection).Doc(ActiveDoc)
}

func (s *Firestore) exclusionsRef(uid string) *firestore.DocumentRef {
	return s.user(uid).Collection(ExclusionsCollection).Doc(ActiveDoc)
}

// Get reads the plan document of the uid.
func (s *Firestore) Get(ctx context.Context, uid string) (Plan, bool, error) {
	if err := checkUID(uid); err != nil {
		return Plan{}, false, err
	}
	snap, err := s.planRef(uid).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return Plan{}, false, nil
	}
	if err != nil {
		return Plan{}, false, err
	}
	var d planDoc
	if err := snap.DataTo(&d); err != nil {
		return Plan{}, false, err
	}
	return d.plan(), true, nil
}

// Exclusions reads the exclusions document of the uid. A uid with no
// document has no exclusion and revision 0.
func (s *Firestore) Exclusions(ctx context.Context, uid string) (Exclusions, error) {
	if err := checkUID(uid); err != nil {
		return Exclusions{}, err
	}
	return decodeExclusions(s.exclusionsRef(uid).Get(ctx))
}

// Save writes both documents in one transaction, after it reads the
// revision of the exclusions.
func (s *Firestore) Save(ctx context.Context, uid string, p Plan, ex Exclusions, replace bool) error {
	if err := checkUID(uid); err != nil {
		return err
	}
	planRef, exRef := s.planRef(uid), s.exclusionsRef(uid)
	return s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		cur, err := decodeExclusions(tx.Get(exRef))
		if err != nil {
			return err
		}
		if cur.Revision != ex.Revision {
			return ErrConflict
		}
		if replace {
			next := ex
			next.Revision = cur.Revision + 1
			if err := tx.Set(exRef, encodeExclusions(next)); err != nil {
				return err
			}
		}
		return tx.Set(planRef, encodePlan(p))
	})
}

func decodeExclusions(snap *firestore.DocumentSnapshot, err error) (Exclusions, error) {
	if status.Code(err) == codes.NotFound {
		return Exclusions{}, nil
	}
	if err != nil {
		return Exclusions{}, err
	}
	var d exclusionsDoc
	if err := snap.DataTo(&d); err != nil {
		return Exclusions{}, err
	}
	out := Exclusions{Revision: d.Revision}
	for _, e := range d.Items {
		out.Items = append(out.Items, Exclusion{Exercise: idOf(e.Exercise), Reason: e.Reason})
	}
	return out, nil
}

func encodeExclusions(x Exclusions) exclusionsDoc {
	d := exclusionsDoc{Revision: x.Revision, Items: []exclusionDoc{}}
	for _, e := range x.Items {
		d.Items = append(d.Items, exclusionDoc{string(e.Exercise), e.Reason})
	}
	return d
}

func (x Exclusions) clone() Exclusions {
	x.Items = slices.Clone(x.Items)
	return x
}
