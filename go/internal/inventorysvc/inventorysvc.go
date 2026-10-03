// Package inventorysvc serves workoutapp.v1.InventoryService (work area
// 4.2). It reads the uid that the auth interceptor stored, and keeps the
// inventory of each uid apart.
package inventorysvc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/inventory"
)

// Server implements workoutappv1connect.InventoryServiceHandler.
type Server struct {
	store   inventory.Store
	catalog domain.Catalog
	newID   func() string
}

var _ workoutappv1connect.InventoryServiceHandler = (*Server)(nil)

// New gives a server over the store, with the product catalog (D-155).
func New(store inventory.Store) *Server {
	return &Server{store: store, catalog: domain.DefaultCatalog(), newID: randomID}
}

// randomID gives a note id of 16 hex characters.
func randomID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func uid(ctx context.Context) (string, error) {
	id := auth.UserID(ctx)
	if id == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))
	}
	return id, nil
}

// fail gives the Connect error of a check or store error. A check error
// names ids and numbers alone, so its text goes to the caller. A store
// error can name a path, so the caller gets a fixed text.
func fail(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, inventory.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, inventory.ErrWeightsChanged):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	return connect.NewError(connect.CodeInternal, errors.New("the inventory store failed"))
}

// GetCatalog gives the product catalog.
func (s *Server) GetCatalog(ctx context.Context, _ *connect.Request[workoutappv1.GetCatalogRequest]) (*connect.Response[workoutappv1.GetCatalogResponse], error) {
	if _, err := uid(ctx); err != nil {
		return nil, err
	}
	out := &workoutappv1.GetCatalogResponse{Version: int32(s.catalog.Version)}
	for _, m := range s.catalog.Machines {
		out.Machines = append(out.Machines, &workoutappv1.CatalogMachine{Id: string(m.ID), Name: m.Name, Kind: string(m.Kind)})
	}
	for _, e := range s.catalog.Exercises {
		out.Exercises = append(out.Exercises, &workoutappv1.CatalogExercise{
			Id: string(e.ID), Name: e.Name, MachineId: string(e.Machine), Region: string(e.Region),
		})
	}
	return connect.NewResponse(out), nil
}

// GetInventory gives the inventory of the caller.
func (s *Server) GetInventory(ctx context.Context, _ *connect.Request[workoutappv1.GetInventoryRequest]) (*connect.Response[workoutappv1.GetInventoryResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	inv, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fail(err)
	}
	return connect.NewResponse(&workoutappv1.GetInventoryResponse{Inventory: s.toProto(inv)}), nil
}

// SaveMachine adds or replaces the entry of a machine (D-193, D-200).
func (s *Server) SaveMachine(ctx context.Context, req *connect.Request[workoutappv1.SaveMachineRequest]) (*connect.Response[workoutappv1.SaveMachineResponse], error) {
	estimates, err := Estimates(req.Msg.GetEstimates())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m := inventory.Machine{Entry: Entry(req.Msg.GetMachineId(), req.Msg.GetWeightsTenthLb(), req.Msg.GetDumbbells()), Estimates: estimates}
	inv, err := s.update(ctx, func(inv inventory.Inventory) (inventory.Inventory, error) {
		return inv.SaveMachine(s.catalog, m)
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workoutappv1.SaveMachineResponse{Inventory: s.toProto(inv)}), nil
}

// ConfirmMachine confirms a machine with the weights that the review
// screen showed (D-193, D-201).
func (s *Server) ConfirmMachine(ctx context.Context, req *connect.Request[workoutappv1.ConfirmMachineRequest]) (*connect.Response[workoutappv1.ConfirmMachineResponse], error) {
	shown := Entry(req.Msg.GetMachineId(), req.Msg.GetWeightsTenthLb(), req.Msg.GetDumbbells())
	inv, err := s.update(ctx, func(inv inventory.Inventory) (inventory.Inventory, error) {
		return inv.ConfirmMachine(s.catalog, shown)
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workoutappv1.ConfirmMachineResponse{Inventory: s.toProto(inv)}), nil
}

// RemoveMachine removes a machine.
func (s *Server) RemoveMachine(ctx context.Context, req *connect.Request[workoutappv1.RemoveMachineRequest]) (*connect.Response[workoutappv1.RemoveMachineResponse], error) {
	id := domain.MachineID(req.Msg.GetMachineId())
	inv, err := s.update(ctx, func(inv inventory.Inventory) (inventory.Inventory, error) {
		return inv.RemoveMachine(id), nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workoutappv1.RemoveMachineResponse{Inventory: s.toProto(inv)}), nil
}

// SaveNote adds or changes a note (D-191).
func (s *Server) SaveNote(ctx context.Context, req *connect.Request[workoutappv1.SaveNoteRequest]) (*connect.Response[workoutappv1.SaveNoteResponse], error) {
	newID := s.newID()
	var noteID string
	inv, err := s.update(ctx, func(inv inventory.Inventory) (inventory.Inventory, error) {
		out, id, err := inv.SaveNote(s.catalog, req.Msg.GetId(), req.Msg.GetText(), newID)
		noteID = id
		return out, err
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workoutappv1.SaveNoteResponse{Inventory: s.toProto(inv), NoteId: noteID}), nil
}

// RemoveNote removes a note.
func (s *Server) RemoveNote(ctx context.Context, req *connect.Request[workoutappv1.RemoveNoteRequest]) (*connect.Response[workoutappv1.RemoveNoteResponse], error) {
	id := req.Msg.GetId()
	inv, err := s.update(ctx, func(inv inventory.Inventory) (inventory.Inventory, error) {
		return inv.RemoveNote(id), nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workoutappv1.RemoveNoteResponse{Inventory: s.toProto(inv)}), nil
}

func (s *Server) update(ctx context.Context, change func(inventory.Inventory) (inventory.Inventory, error)) (inventory.Inventory, error) {
	id, err := uid(ctx)
	if err != nil {
		return inventory.Inventory{}, err
	}
	inv, err := s.store.Update(ctx, id, change)
	if err != nil {
		return inventory.Inventory{}, fail(err)
	}
	return inv, nil
}

// ErrTwoEstimates is the check error of two estimates for one exercise.
var ErrTwoEstimates = fmt.Errorf("%w: two estimates for one exercise", domain.ErrInvalid)

// Estimates gives the estimates of the contract by exercise, or nil for
// none. It refuses two estimates for one exercise. Machine.Check reads
// each load.
func Estimates(list []*workoutappv1.Estimate) (map[domain.ExerciseID]domain.Load, error) {
	var out map[domain.ExerciseID]domain.Load
	for _, e := range list {
		ex := domain.ExerciseID(e.GetExerciseId())
		if _, dup := out[ex]; dup {
			return nil, ErrTwoEstimates
		}
		if out == nil {
			out = map[domain.ExerciseID]domain.Load{}
		}
		out[ex] = domain.Load(e.GetLoadTenthLb())
	}
	return out, nil
}

// Entry gives the inventory entry of a machine of the contract: its
// catalog id, and its weights or its dumbbell set.
func Entry(machine string, weights []int32, d *workoutappv1.DumbbellSet) domain.InventoryEntry {
	e := domain.InventoryEntry{Machine: domain.MachineID(machine)}
	for _, w := range weights {
		e.Weights = append(e.Weights, domain.Load(w))
	}
	if d != nil {
		e.Dumbbells = &domain.DumbbellSet{
			Lightest: domain.Load(d.GetLightestTenthLb()),
			Heaviest: domain.Load(d.GetHeaviestTenthLb()),
			Step:     domain.Load(d.GetStepTenthLb()),
		}
	}
	return e
}

// toProto gives the inventory of the contract. Each int32 holds its load:
// a stored load passed the bounds of D-199, so it is 10,000 tenths or
// less.
func (s *Server) toProto(inv inventory.Inventory) *workoutappv1.Inventory {
	out := &workoutappv1.Inventory{}
	for _, m := range inv.Machines {
		pm := &workoutappv1.InventoryMachine{MachineId: string(m.Entry.Machine), State: state(m.State)}
		for _, w := range m.Entry.Weights {
			pm.WeightsTenthLb = append(pm.WeightsTenthLb, int32(w))
		}
		if d := m.Entry.Dumbbells; d != nil {
			pm.Dumbbells = &workoutappv1.DumbbellSet{
				LightestTenthLb: int32(d.Lightest), HeaviestTenthLb: int32(d.Heaviest), StepTenthLb: int32(d.Step),
			}
		}
		for _, e := range s.catalog.ExercisesOnMachine(m.Entry.Machine) {
			if load, ok := m.Estimates[e.ID]; ok {
				pm.Estimates = append(pm.Estimates, &workoutappv1.Estimate{ExerciseId: string(e.ID), LoadTenthLb: int32(load)})
			}
		}
		out.Machines = append(out.Machines, pm)
	}
	for _, n := range inv.Notes {
		out.Notes = append(out.Notes, &workoutappv1.InventoryNote{Id: n.ID, Text: n.Text})
	}
	return out
}

func state(s inventory.State) workoutappv1.MachineState {
	switch s {
	case inventory.Draft:
		return workoutappv1.MachineState_MACHINE_STATE_DRAFT
	case inventory.Confirmed:
		return workoutappv1.MachineState_MACHINE_STATE_CONFIRMED
	}
	return workoutappv1.MachineState_MACHINE_STATE_UNSPECIFIED
}
