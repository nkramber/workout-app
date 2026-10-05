package usersvc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/plan"
)

func TestGetMe(t *testing.T) {
	req := connect.NewRequest(&workoutappv1.GetMeRequest{})
	res, err := Server{}.GetMe(auth.WithUserID(context.Background(), "uid-a"), req)
	if err != nil || res.Msg.GetUid() != "uid-a" {
		t.Fatalf("GetMe = %v, %v, want uid-a", res, err)
	}
	if _, err := (Server{}).GetMe(context.Background(), req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("GetMe with no uid = %v, want Unauthenticated", err)
	}
}

// fakeWorkouts and fakePlans record the deletes, and give err.
type fakeWorkouts struct {
	n     int
	calls []string
	err   error
}

func (f *fakeWorkouts) DeleteAll(_ context.Context, uid string) (int, error) {
	f.calls = append(f.calls, uid)
	if f.err != nil {
		return 0, f.err
	}
	n := f.n
	f.n = 0
	return n, nil
}

type fakePlans struct {
	calls []string
	err   error
}

func (f *fakePlans) Delete(_ context.Context, uid string) error {
	f.calls = append(f.calls, uid)
	return f.err
}

// TestDeleteHistory: the call needs the text of D-314, and deletes the
// workouts, the plan, and the error records of the caller alone (D-315).
// A second call passes. A failed step gives UNAVAILABLE with no uid and
// no path, and a later call deletes the rest.
func TestDeleteHistory(t *testing.T) {
	ctx := auth.WithUserID(context.Background(), "uid-a")
	errs := &plan.MemoryErrors{}
	for _, u := range []string{"uid-a", "uid-b", "uid-a"} {
		if err := errs.Add(ctx, plan.ErrorRecord{User: u}); err != nil {
			t.Fatal(err)
		}
	}
	w, p := &fakeWorkouts{n: 3}, &fakePlans{}
	s := Server{History: &History{Workouts: w, Plans: p, Errors: errs}}
	call := func(ctx context.Context, text string) (*connect.Response[workoutappv1.DeleteHistoryResponse], error) {
		return s.DeleteHistory(ctx, connect.NewRequest(&workoutappv1.DeleteHistoryRequest{Confirmation: text}))
	}

	for _, text := range []string{"", "delete all data", "Delete all data ", "Delete all"} {
		if _, err := call(ctx, text); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Fatalf("confirmation %q: %v, want InvalidArgument", text, err)
		}
	}
	if len(w.calls)+len(p.calls) != 0 || len(errs.Records()) != 3 {
		t.Fatal("a refused confirmation deleted data")
	}
	if _, err := call(context.Background(), Confirmation); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("no uid: %v, want Unauthenticated", err)
	}

	res, err := call(ctx, Confirmation)
	if err != nil || res.Msg.GetDeletedWorkouts() != 3 {
		t.Fatalf("DeleteHistory = %v, %v, want 3 workouts", res, err)
	}
	if !slices.Equal(w.calls, []string{"uid-a"}) || !slices.Equal(p.calls, []string{"uid-a"}) {
		t.Fatalf("deletes %v and %v, want uid-a", w.calls, p.calls)
	}
	if got := errs.Records(); len(got) != 1 || got[0].User != "uid-b" {
		t.Fatalf("error records %+v, want the record of uid-b alone", got)
	}
	if res, err := call(ctx, Confirmation); err != nil || res.Msg.GetDeletedWorkouts() != 0 {
		t.Fatalf("second call = %v, %v, want 0 workouts", res, err)
	}

	p.err = errors.New("rpc error: users/uid-a/plan/active: unavailable")
	_, err = call(ctx, Confirmation)
	if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "uid-a") || strings.Contains(err.Error(), "users/") {
		t.Fatalf("a failed delete: %v, want UNAVAILABLE with no uid and no path", err)
	}
	if _, err := (Server{}).DeleteHistory(ctx, connect.NewRequest(&workoutappv1.DeleteHistoryRequest{Confirmation: Confirmation})); connect.CodeOf(err) != connect.CodeUnimplemented {
		t.Fatalf("no stores: %v, want Unimplemented", err)
	}
}
