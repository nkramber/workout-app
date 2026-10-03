package inventory

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OpsCollection holds one document for each applied op id of a user, at
// users/{uid}/ops/{opId}. The workout entries of the outbox use the same
// collection, so one op id applies one time, whatever its entity (D-256,
// D-257).
const OpsCollection = "ops"

// Op names an outbox entry that changes the inventory (D-272). The caller
// checks the op id first, because it becomes a document id.
type Op struct {
	ID          string
	Entity      string
	EntityID    string
	BaseVersion int64
	At          time.Time
}

// ApplyOp applies the change under the lock of the store.
func (s *Memory) ApplyOp(_ context.Context, uid string, op Op, change func(Inventory) (Inventory, error)) (bool, error) {
	if err := checkUID(uid); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ops[uid][op.ID] {
		return true, nil
	}
	next, err := change(s.data[uid].clone())
	if err != nil {
		return false, err
	}
	s.data[uid] = next.clone()
	if s.ops == nil {
		s.ops = map[string]map[string]bool{}
	}
	if s.ops[uid] == nil {
		s.ops[uid] = map[string]bool{}
	}
	s.ops[uid][op.ID] = true
	return false, nil
}

// ApplyOp runs one transaction: it reads the op id and the inventory,
// then writes the inventory and the op id. Firestore runs the function
// again when another transaction changes a document that it read, so two
// calls with the same op id apply it one time.
func (s *Firestore) ApplyOp(ctx context.Context, uid string, op Op, change func(Inventory) (Inventory, error)) (bool, error) {
	if err := checkUID(uid); err != nil {
		return false, err
	}
	user := s.client.Collection(UsersCollection).Doc(uid)
	opRef := user.Collection(OpsCollection).Doc(op.ID)
	ref := s.ref(uid)
	var replayed bool
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		replayed = false
		_, err := tx.Get(opRef)
		switch {
		case err == nil:
			replayed = true
			return nil
		case status.Code(err) != codes.NotFound:
			return err
		}
		cur, err := decode(tx.Get(ref))
		if err != nil {
			return err
		}
		next, err := change(cur)
		if err != nil {
			return err
		}
		if err := tx.Set(ref, encode(next)); err != nil {
			return err
		}
		return tx.Create(opRef, opDoc{
			Entity: op.Entity, EntityID: op.EntityID, BaseVersion: op.BaseVersion,
			At: op.At.UTC(), AppliedAt: time.Now().UTC(),
		})
	}, firestore.MaxAttempts(5))
	if err != nil {
		return false, err
	}
	return replayed, nil
}

// The stored op id of an inventory entry. It holds ids and times alone,
// in the form of the op id document of the workout store. The inventory
// has no version, so the version is 0.
type opDoc struct {
	Entity      string    `firestore:"entity"`
	EntityID    string    `firestore:"entity_id"`
	BaseVersion int64     `firestore:"base_version"`
	Version     int64     `firestore:"version"`
	At          time.Time `firestore:"at"`
	AppliedAt   time.Time `firestore:"applied_at"`
}
