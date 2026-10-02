package profile

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
		if _, _, err := s.Get(ctx, uid); !errors.Is(err, errUID) {
			t.Errorf("Get(%q) = %v, want the uid error", uid, err)
		}
		if err := s.Save(ctx, uid, valid()); !errors.Is(err, errUID) {
			t.Errorf("Save(%q) = %v, want the uid error", uid, err)
		}
	}
	if _, ok, err := s.Get(ctx, "uid-a"); ok || err != nil {
		t.Fatalf("Get of a new uid = %v, %v, want no profile", ok, err)
	}
	p := valid()
	if err := s.Save(ctx, "uid-a", p); err != nil {
		t.Fatal(err)
	}
	p.Groups[0] = "back"
	got, ok, err := s.Get(ctx, "uid-a")
	if err != nil || !ok || !reflect.DeepEqual(got, valid()) {
		t.Fatalf("Get = %+v, %v, %v, want the saved profile", got, ok, err)
	}
	if _, ok, _ := s.Get(ctx, "uid-b"); ok {
		t.Fatal("uid-b reads the profile of uid-a")
	}
}
