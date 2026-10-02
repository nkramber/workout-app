package inventory

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemory()
	for _, uid := range []string{"", "a/b"} {
		if _, err := s.Get(ctx, uid); !errors.Is(err, errUID) {
			t.Errorf("Get(%q) = %v, want the uid error", uid, err)
		}
	}
	saved, err := s.Update(ctx, "uid-a", func(inv Inventory) (Inventory, error) { return inv.SaveMachine(catalog, legPress(0)) })
	if err != nil {
		t.Fatal(err)
	}
	refused := errors.New("refused")
	if _, err := s.Update(ctx, "uid-a", func(inv Inventory) (Inventory, error) {
		inv.Machines[0].State = Confirmed
		return inv, refused
	}); err != refused {
		t.Fatalf("Update = %v, want the error of the change as it is", err)
	}
	got, err := s.Get(ctx, "uid-a")
	if err != nil || !reflect.DeepEqual(got, saved) {
		t.Fatalf("after a refused change: %+v, %v, want %+v", got, err, saved)
	}
	if other, _ := s.Get(ctx, "uid-b"); len(other.Machines) != 0 {
		t.Fatalf("uid-b reads %+v", other)
	}
}
