// Package profilesvc serves workoutapp.v1.ProfileService (work area 5.1).
// It reads the uid that the auth interceptor stored, and keeps the
// profile of each uid apart.
package profilesvc

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/profile"
)

// Server implements workoutappv1connect.ProfileServiceHandler.
type Server struct {
	store   profile.Store
	catalog domain.Catalog
	tables  domain.BodyTables
}

var _ workoutappv1connect.ProfileServiceHandler = (*Server)(nil)

// New gives a server over the store, with the product catalog (D-155)
// and the body tables (D-208, D-210).
func New(store profile.Store) *Server {
	return &Server{store: store, catalog: domain.DefaultCatalog(), tables: domain.DefaultBodyTables()}
}

func uid(ctx context.Context) (string, error) {
	id := auth.UserID(ctx)
	if id == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))
	}
	return id, nil
}

// fail gives the Connect error of a check or store error. A check error
// names the field, ids, and bounds alone, so its text goes to the
// caller. A store error can name a path, so the caller gets a fixed
// text.
func fail(err error) error {
	if errors.Is(err, domain.ErrInvalid) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewError(connect.CodeInternal, errors.New("the profile store failed"))
}

// GetProfileOptions gives the fixed lists of the profile.
func (s *Server) GetProfileOptions(ctx context.Context, _ *connect.Request[workoutappv1.GetProfileOptionsRequest]) (*connect.Response[workoutappv1.GetProfileOptionsResponse], error) {
	if _, err := uid(ctx); err != nil {
		return nil, err
	}
	out := &workoutappv1.GetProfileOptionsResponse{}
	for _, e := range profile.Experiences() {
		out.Experiences = append(out.Experiences, &workoutappv1.Option{Id: string(e), Name: e.Name()})
	}
	for _, t := range s.tables.Templates {
		out.GoalTemplates = append(out.GoalTemplates, &workoutappv1.GoalTemplate{
			Id: string(t.ID), Name: t.Name, MuscleGroups: strs(t.Groups),
		})
	}
	for _, g := range domain.MuscleGroups() {
		out.MuscleGroups = append(out.MuscleGroups, &workoutappv1.Option{Id: string(g), Name: g.Name()})
	}
	for _, a := range domain.Areas() {
		out.InjuryAreas = append(out.InjuryAreas, &workoutappv1.Option{Id: string(a), Name: a.Name()})
	}
	return connect.NewResponse(out), nil
}

// GetProfile gives the profile of the caller, or no profile.
func (s *Server) GetProfile(ctx context.Context, _ *connect.Request[workoutappv1.GetProfileRequest]) (*connect.Response[workoutappv1.GetProfileResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	p, ok, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fail(err)
	}
	out := &workoutappv1.GetProfileResponse{}
	if ok {
		out.Profile = toProto(p)
	}
	return connect.NewResponse(out), nil
}

// SaveProfile checks the profile and replaces the stored one.
func (s *Server) SaveProfile(ctx context.Context, req *connect.Request[workoutappv1.SaveProfileRequest]) (*connect.Response[workoutappv1.SaveProfileResponse], error) {
	id, err := uid(ctx)
	if err != nil {
		return nil, err
	}
	in := req.Msg.GetProfile()
	if in == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("a profile is required"))
	}
	p := fromProto(in).Normalize(s.catalog)
	if err := p.Check(s.catalog, s.tables); err != nil {
		return nil, fail(err)
	}
	if err := s.store.Save(ctx, id, p); err != nil {
		return nil, fail(err)
	}
	return connect.NewResponse(&workoutappv1.SaveProfileResponse{Profile: toProto(p)}), nil
}

func fromProto(in *workoutappv1.Profile) profile.Profile {
	return profile.Profile{
		Experience:   profile.Experience(in.GetExperience()),
		Template:     domain.TemplateID(in.GetGoalTemplate()),
		Groups:       ids[domain.MuscleGroup](in.GetMuscleGroups()),
		FreeText:     in.GetFreeText(),
		InjuredAreas: ids[domain.Area](in.GetInjuredAreas()),
		InjuryText:   in.GetInjuryText(),
		AgeYears:     int(in.GetAgeYears()),
		HeightIn:     int(in.GetHeightIn()),
		WeightLb:     int(in.GetWeightLb()),
		Cardio:       ids[domain.ExerciseID](in.GetCardioExerciseIds()),
		TrainingDays: int(in.GetTrainingDays()),
	}
}

// toProto gives the profile of the contract. Each number of a stored
// profile passed its bound, so it fits an int32.
func toProto(p profile.Profile) *workoutappv1.Profile {
	return &workoutappv1.Profile{
		Experience:        string(p.Experience),
		GoalTemplate:      string(p.Template),
		MuscleGroups:      strs(p.Groups),
		FreeText:          p.FreeText,
		InjuredAreas:      strs(p.InjuredAreas),
		InjuryText:        p.InjuryText,
		AgeYears:          int32(p.AgeYears),
		HeightIn:          int32(p.HeightIn),
		WeightLb:          int32(p.WeightLb),
		CardioExerciseIds: strs(p.Cardio),
		TrainingDays:      int32(p.TrainingDays),
	}
}

func ids[T ~string](in []string) []T {
	var out []T
	for _, v := range in {
		out = append(out, T(v))
	}
	return out
}

func strs[T ~string](in []T) []string {
	var out []string
	for _, v := range in {
		out = append(out, string(v))
	}
	return out
}
