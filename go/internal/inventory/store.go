package inventory

import (
	"context"
	"errors"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// The Firestore path of the one active inventory of a user is
// users/{uid}/inventory/active (D-197). One document holds each machine
// and each note, so one transaction reads and writes the whole
// inventory, and the check of one entry for each machine holds.
const (
	UsersCollection = "users"
	Collection      = "inventory"
	ActiveDoc       = "active"
)

// Store reads and changes the inventory of a user. Update runs change on
// the stored inventory and stores the result as one atomic step. A store
// can run change more than one time, so change must have no side effect.
// An error of change comes back from Update as it is.
//
// ApplyOp applies an inventory entry of the outbox (D-272). It reads the
// op id first. An applied op id changes nothing, and gives replayed true.
// Else ApplyOp runs change as Update does, and stores the result and the
// op id together, or stores nothing. After an error of change, the op id
// stays unapplied, so a later entry with the same op id gets the same
// check.
type Store interface {
	Get(ctx context.Context, uid string) (Inventory, error)
	Update(ctx context.Context, uid string, change func(Inventory) (Inventory, error)) (Inventory, error)
	ApplyOp(ctx context.Context, uid string, op Op, change func(Inventory) (Inventory, error)) (replayed bool, err error)
}

var (
	_ Store = (*Memory)(nil)
	_ Store = (*Firestore)(nil)
)

var errUID = errors.New("inventory: a uid of 1 or more characters with no slash is required")

func checkUID(uid string) error {
	if uid == "" || strings.Contains(uid, "/") {
		return errUID
	}
	return nil
}

// Memory is a Store in memory, for tests. ops holds the applied op ids
// of each uid.
type Memory struct {
	mu   sync.Mutex
	data map[string]Inventory
	ops  map[string]map[string]bool
}

// NewMemory gives an empty Memory store.
func NewMemory() *Memory { return &Memory{data: map[string]Inventory{}} }

// Get gives the inventory of the uid, or an empty one.
func (s *Memory) Get(_ context.Context, uid string) (Inventory, error) {
	if err := checkUID(uid); err != nil {
		return Inventory{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[uid].clone(), nil
}

// Update runs change under the lock of the store.
func (s *Memory) Update(_ context.Context, uid string, change func(Inventory) (Inventory, error)) (Inventory, error) {
	if err := checkUID(uid); err != nil {
		return Inventory{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next, err := change(s.data[uid].clone())
	if err != nil {
		return Inventory{}, err
	}
	s.data[uid] = next.clone()
	return next, nil
}

// Firestore is the production Store.
type Firestore struct {
	client *firestore.Client
}

// FromFirestore gives the Store of the client.
func FromFirestore(client *firestore.Client) *Firestore { return &Firestore{client: client} }

func (s *Firestore) ref(uid string) *firestore.DocumentRef {
	return s.client.Collection(UsersCollection).Doc(uid).Collection(Collection).Doc(ActiveDoc)
}

// Get reads the document of the uid. A uid with no document has an
// empty inventory.
func (s *Firestore) Get(ctx context.Context, uid string) (Inventory, error) {
	if err := checkUID(uid); err != nil {
		return Inventory{}, err
	}
	snap, err := s.ref(uid).Get(ctx)
	return decode(snap, err)
}

// Update reads, changes, and writes the document in one transaction.
func (s *Firestore) Update(ctx context.Context, uid string, change func(Inventory) (Inventory, error)) (Inventory, error) {
	if err := checkUID(uid); err != nil {
		return Inventory{}, err
	}
	ref := s.ref(uid)
	var next Inventory
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		cur, err := decode(tx.Get(ref))
		if err != nil {
			return err
		}
		if next, err = change(cur); err != nil {
			return err
		}
		return tx.Set(ref, encode(next))
	})
	if err != nil {
		return Inventory{}, err
	}
	return next, nil
}

// The stored document. Each load is a whole number of tenths of a pound,
// as domain.Load. A machine holds its catalog id, its weights, its
// estimates, and its state alone (D-54, D-192, D-193).
type document struct {
	Machines []machineDoc `firestore:"machines"`
	Notes    []noteDoc    `firestore:"notes"`
}

type machineDoc struct {
	Machine   string           `firestore:"machine"`
	Weights   []int64          `firestore:"weights,omitempty"`
	Dumbbells *dumbbellDoc     `firestore:"dumbbells,omitempty"`
	Estimates map[string]int64 `firestore:"estimates,omitempty"`
	State     string           `firestore:"state"`
}

type dumbbellDoc struct {
	Lightest int64 `firestore:"lightest"`
	Heaviest int64 `firestore:"heaviest"`
	Step     int64 `firestore:"step"`
}

type noteDoc struct {
	ID   string `firestore:"id"`
	Text string `firestore:"text"`
}

func decode(snap *firestore.DocumentSnapshot, err error) (Inventory, error) {
	if status.Code(err) == codes.NotFound {
		return Inventory{}, nil
	}
	if err != nil {
		return Inventory{}, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return Inventory{}, err
	}
	var inv Inventory
	for _, md := range d.Machines {
		m := Machine{Entry: domain.InventoryEntry{Machine: domain.MachineID(md.Machine)}, State: State(md.State)}
		for _, w := range md.Weights {
			m.Entry.Weights = append(m.Entry.Weights, domain.Load(w))
		}
		if md.Dumbbells != nil {
			m.Entry.Dumbbells = &domain.DumbbellSet{
				Lightest: domain.Load(md.Dumbbells.Lightest),
				Heaviest: domain.Load(md.Dumbbells.Heaviest),
				Step:     domain.Load(md.Dumbbells.Step),
			}
		}
		for ex, load := range md.Estimates {
			if m.Estimates == nil {
				m.Estimates = map[domain.ExerciseID]domain.Load{}
			}
			m.Estimates[domain.ExerciseID(ex)] = domain.Load(load)
		}
		inv.Machines = append(inv.Machines, m)
	}
	for _, nd := range d.Notes {
		inv.Notes = append(inv.Notes, Note{ID: nd.ID, Text: nd.Text})
	}
	return inv, nil
}

func encode(inv Inventory) document {
	d := document{Machines: []machineDoc{}, Notes: []noteDoc{}}
	for _, m := range inv.Machines {
		md := machineDoc{Machine: string(m.Entry.Machine), State: string(m.State)}
		for _, w := range m.Entry.Weights {
			md.Weights = append(md.Weights, int64(w))
		}
		if ds := m.Entry.Dumbbells; ds != nil {
			md.Dumbbells = &dumbbellDoc{Lightest: int64(ds.Lightest), Heaviest: int64(ds.Heaviest), Step: int64(ds.Step)}
		}
		for ex, load := range m.Estimates {
			if md.Estimates == nil {
				md.Estimates = map[string]int64{}
			}
			md.Estimates[string(ex)] = int64(load)
		}
		d.Machines = append(d.Machines, md)
	}
	for _, n := range inv.Notes {
		d.Notes = append(d.Notes, noteDoc{ID: n.ID, Text: n.Text})
	}
	return d
}
