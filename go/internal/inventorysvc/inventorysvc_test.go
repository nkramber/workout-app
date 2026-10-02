package inventorysvc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
)

var signedIn = auth.WithUserID(context.Background(), "uid-a")

func save(t *testing.T, s *Server, req *workoutappv1.SaveMachineRequest) (*workoutappv1.Inventory, error) {
	t.Helper()
	res, err := s.SaveMachine(signedIn, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return res.Msg.GetInventory(), nil
}

func legPress() *workoutappv1.SaveMachineRequest {
	return &workoutappv1.SaveMachineRequest{
		MachineId:      "leg_press",
		WeightsTenthLb: []int32{200, 300, 400, 500},
		Estimates:      []*workoutappv1.Estimate{{ExerciseId: "leg_press", LoadTenthLb: 300}},
	}
}

func TestNeedsAUid(t *testing.T) {
	s := New(inventory.NewMemory())
	ctx := context.Background()
	if _, err := s.GetCatalog(ctx, connect.NewRequest(&workoutappv1.GetCatalogRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetCatalog = %v", err)
	}
	if _, err := s.GetInventory(ctx, connect.NewRequest(&workoutappv1.GetInventoryRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetInventory = %v", err)
	}
	if _, err := s.SaveNote(ctx, connect.NewRequest(&workoutappv1.SaveNoteRequest{Text: "x"})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("SaveNote = %v", err)
	}
}

func TestGetCatalog(t *testing.T) {
	res, err := New(inventory.NewMemory()).GetCatalog(signedIn, connect.NewRequest(&workoutappv1.GetCatalogRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	c := domain.DefaultCatalog()
	got := res.Msg
	if int(got.GetVersion()) != c.Version || len(got.GetMachines()) != len(c.Machines) || len(got.GetExercises()) != len(c.Exercises) {
		t.Fatalf("catalog = version %d, %d machines, %d exercises", got.GetVersion(), len(got.GetMachines()), len(got.GetExercises()))
	}
	m, e := got.GetMachines()[13], got.GetExercises()[0]
	if m.GetId() != "dumbbells" || m.GetKind() != "dumbbell" || e.GetMachineId() != "leg_press" || e.GetRegion() != "lower_push" {
		t.Fatalf("catalog items = %v, %v", m, e)
	}
}

func TestSaveConfirmRemove(t *testing.T) {
	s := New(inventory.NewMemory())
	inv, err := save(t, s, legPress())
	if err != nil {
		t.Fatal(err)
	}
	m := inv.GetMachines()[0]
	if m.GetState() != workoutappv1.MachineState_MACHINE_STATE_DRAFT || m.GetEstimates()[0].GetLoadTenthLb() != 300 {
		t.Fatalf("saved machine = %v", m)
	}

	confirm := func(weights []int32) (*connect.Response[workoutappv1.ConfirmMachineResponse], error) {
		return s.ConfirmMachine(signedIn, connect.NewRequest(&workoutappv1.ConfirmMachineRequest{MachineId: "leg_press", WeightsTenthLb: weights}))
	}
	if _, err := confirm([]int32{200, 300}); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("confirm with other weights = %v, want FailedPrecondition", err)
	}
	res, err := confirm(legPress().GetWeightsTenthLb())
	if err != nil || res.Msg.GetInventory().GetMachines()[0].GetState() != workoutappv1.MachineState_MACHINE_STATE_CONFIRMED {
		t.Fatalf("confirm = %v, %v", res, err)
	}
	if _, err := s.ConfirmMachine(signedIn, connect.NewRequest(&workoutappv1.ConfirmMachineRequest{MachineId: "chest_press"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("confirm of an unknown machine = %v, want NotFound", err)
	}

	for range 2 {
		rm, err := s.RemoveMachine(signedIn, connect.NewRequest(&workoutappv1.RemoveMachineRequest{MachineId: "leg_press"}))
		if err != nil || len(rm.Msg.GetInventory().GetMachines()) != 0 {
			t.Fatalf("remove = %v, %v", rm, err)
		}
	}
}

func TestSaveMachineRefuses(t *testing.T) {
	s := New(inventory.NewMemory())
	dup := legPress()
	dup.Estimates = append(dup.Estimates, &workoutappv1.Estimate{ExerciseId: "leg_press", LoadTenthLb: 400})
	high := legPress()
	high.Estimates[0].LoadTenthLb = 600
	cases := map[string]*workoutappv1.SaveMachineRequest{
		"unknown machine":         {MachineId: "barbell"},
		"empty machine id":        {},
		"two estimates":           dup,
		"estimate above the list": high,
		"weight above the bound":  {MachineId: "leg_press", WeightsTenthLb: []int32{10010}},
		"negative weight":         {MachineId: "leg_press", WeightsTenthLb: []int32{-10}},
		"dumbbell set on a stack": {MachineId: "leg_press", WeightsTenthLb: []int32{100}, Dumbbells: &workoutappv1.DumbbellSet{LightestTenthLb: 50, HeaviestTenthLb: 100, StepTenthLb: 50}},
	}
	for name, req := range cases {
		if _, err := save(t, s, req); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: SaveMachine = %v, want InvalidArgument", name, err)
		}
	}
}

func TestDumbbellsAndEstimateOrder(t *testing.T) {
	s := New(inventory.NewMemory())
	inv, err := save(t, s, &workoutappv1.SaveMachineRequest{
		MachineId: "dumbbells",
		Dumbbells: &workoutappv1.DumbbellSet{LightestTenthLb: 50, HeaviestTenthLb: 500, StepTenthLb: 25},
		Estimates: []*workoutappv1.Estimate{
			{ExerciseId: "db_goblet_squat", LoadTenthLb: 400},
			{ExerciseId: "db_flat_bench_press", LoadTenthLb: 325},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	m := inv.GetMachines()[0]
	if d := m.GetDumbbells(); d.GetLightestTenthLb() != 50 || d.GetHeaviestTenthLb() != 500 || d.GetStepTenthLb() != 25 || len(m.GetWeightsTenthLb()) != 0 {
		t.Fatalf("dumbbells = %v", m)
	}
	if es := m.GetEstimates(); len(es) != 2 || es[0].GetExerciseId() != "db_flat_bench_press" || es[1].GetExerciseId() != "db_goblet_squat" {
		t.Fatalf("estimates = %v, want catalog order", es)
	}
}

func TestNotes(t *testing.T) {
	s := New(inventory.NewMemory())
	res, err := s.SaveNote(signedIn, connect.NewRequest(&workoutappv1.SaveNoteRequest{Text: " Hack squat "}))
	if err != nil {
		t.Fatal(err)
	}
	id := res.Msg.GetNoteId()
	if len(id) != 16 || res.Msg.GetInventory().GetNotes()[0].GetText() != "Hack squat" {
		t.Fatalf("SaveNote = %v", res.Msg)
	}
	if _, err := s.SaveNote(signedIn, connect.NewRequest(&workoutappv1.SaveNoteRequest{Id: "unknown", Text: "x"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("change of an unknown note = %v, want NotFound", err)
	}
	if _, err := s.SaveNote(signedIn, connect.NewRequest(&workoutappv1.SaveNoteRequest{Text: strings.Repeat("x", 201)})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("long note = %v, want InvalidArgument", err)
	}
	rm, err := s.RemoveNote(signedIn, connect.NewRequest(&workoutappv1.RemoveNoteRequest{Id: id}))
	if err != nil || len(rm.Msg.GetInventory().GetNotes()) != 0 {
		t.Fatalf("RemoveNote = %v, %v", rm, err)
	}
}

type brokenStore struct{}

func (brokenStore) Get(context.Context, string) (inventory.Inventory, error) {
	return inventory.Inventory{}, errors.New("rpc error: projects/p/databases/(default)/documents/users/uid-a")
}

func (brokenStore) Update(context.Context, string, func(inventory.Inventory) (inventory.Inventory, error)) (inventory.Inventory, error) {
	return inventory.Inventory{}, errors.New("rpc error: projects/p/databases/(default)/documents/users/uid-a")
}

// TestStoreErrorIsHidden gives the caller a fixed text for a store
// error, so no path or uid leaves the server.
func TestStoreErrorIsHidden(t *testing.T) {
	s := New(brokenStore{})
	_, err := s.GetInventory(signedIn, connect.NewRequest(&workoutappv1.GetInventoryRequest{}))
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users") {
		t.Fatalf("GetInventory = %v, want Internal with no path", err)
	}
	_, err = save(t, s, legPress())
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users") {
		t.Fatalf("SaveMachine = %v, want Internal with no path", err)
	}
}
