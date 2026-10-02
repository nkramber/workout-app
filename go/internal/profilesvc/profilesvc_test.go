package profilesvc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
	"github.com/nkramber/workout-app/go/internal/profile"
)

var signedIn = auth.WithUserID(context.Background(), "uid-a")

// synthetic gives a valid profile of the contract, with synthetic data.
func synthetic() *workoutappv1.Profile {
	return &workoutappv1.Profile{
		Experience:        "intermediate",
		GoalTemplate:      "strength",
		MuscleGroups:      []string{"back", "chest"},
		FreeText:          " More upper body. ",
		InjuredAreas:      []string{"knee"},
		InjuryText:        "Old knee strain.",
		AgeYears:          32,
		HeightIn:          70,
		WeightLb:          180,
		CardioExerciseIds: []string{"rowing_machine", "upright_bike"},
		TrainingDays:      3,
	}
}

func save(s *Server, p *workoutappv1.Profile) (*workoutappv1.Profile, error) {
	res, err := s.SaveProfile(signedIn, connect.NewRequest(&workoutappv1.SaveProfileRequest{Profile: p}))
	if err != nil {
		return nil, err
	}
	return res.Msg.GetProfile(), nil
}

func TestNeedsAUid(t *testing.T) {
	s := New(profile.NewMemory())
	ctx := context.Background()
	if _, err := s.GetProfileOptions(ctx, connect.NewRequest(&workoutappv1.GetProfileOptionsRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetProfileOptions = %v", err)
	}
	if _, err := s.GetProfile(ctx, connect.NewRequest(&workoutappv1.GetProfileRequest{})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetProfile = %v", err)
	}
	if _, err := s.SaveProfile(ctx, connect.NewRequest(&workoutappv1.SaveProfileRequest{Profile: synthetic()})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("SaveProfile = %v", err)
	}
}

func TestGetProfileOptions(t *testing.T) {
	res, err := New(profile.NewMemory()).GetProfileOptions(signedIn, connect.NewRequest(&workoutappv1.GetProfileOptionsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	ids := func(opts []*workoutappv1.Option) []string {
		var out []string
		for _, o := range opts {
			if o.GetName() == "" {
				t.Errorf("option %q has no name", o.GetId())
			}
			out = append(out, o.GetId())
		}
		return out
	}
	m := res.Msg
	if got := ids(m.GetExperiences()); !slices.Equal(got, []string{"intermediate", "advanced"}) {
		t.Errorf("experiences = %v", got)
	}
	if got := ids(m.GetInjuryAreas()); !slices.Equal(got, []string{"shoulder", "elbow", "wrist", "lower_back", "hip", "knee", "ankle"}) {
		t.Errorf("areas = %v", got)
	}
	if got := ids(m.GetMuscleGroups()); len(got) != 10 || got[0] != "chest" || got[9] != "core" {
		t.Errorf("groups = %v", got)
	}
	ts := m.GetGoalTemplates()
	if len(ts) != 2 || ts[0].GetId() != "general_fitness" || len(ts[0].GetMuscleGroups()) != 10 || ts[1].GetId() != "strength" {
		t.Errorf("templates = %v", ts)
	}
}

func TestSaveAndGet(t *testing.T) {
	s := New(profile.NewMemory())
	res, err := s.GetProfile(signedIn, connect.NewRequest(&workoutappv1.GetProfileRequest{}))
	if err != nil || res.Msg.GetProfile() != nil {
		t.Fatalf("GetProfile before a save = %v, %v, want no profile", res, err)
	}
	saved, err := save(s, synthetic())
	if err != nil {
		t.Fatal(err)
	}
	// The server trims each text and orders each list.
	if saved.GetFreeText() != "More upper body." ||
		!slices.Equal(saved.GetMuscleGroups(), []string{"chest", "back"}) ||
		!slices.Equal(saved.GetCardioExerciseIds(), []string{"upright_bike", "rowing_machine"}) {
		t.Fatalf("saved = %v", saved)
	}
	res, err = s.GetProfile(signedIn, connect.NewRequest(&workoutappv1.GetProfileRequest{}))
	if err != nil || res.Msg.GetProfile().String() != saved.String() {
		t.Fatalf("GetProfile = %v, %v, want %v", res, err, saved)
	}
}

func TestSaveRefuses(t *testing.T) {
	s := New(profile.NewMemory())
	cases := map[string]func(*workoutappv1.Profile){
		"age 17":           func(p *workoutappv1.Profile) { p.AgeYears = 17 },
		"height 97":        func(p *workoutappv1.Profile) { p.HeightIn = 97 },
		"weight 501":       func(p *workoutappv1.Profile) { p.WeightLb = 501 },
		"5 training days":  func(p *workoutappv1.Profile) { p.TrainingDays = 5 },
		"beginner":         func(p *workoutappv1.Profile) { p.Experience = "beginner" },
		"unknown template": func(p *workoutappv1.Profile) { p.GoalTemplate = "bulk" },
		"unknown group":    func(p *workoutappv1.Profile) { p.MuscleGroups = []string{"neck"} },
		"unknown area":     func(p *workoutappv1.Profile) { p.InjuredAreas = []string{"neck"} },
		"leg press cardio": func(p *workoutappv1.Profile) { p.CardioExerciseIds = []string{"leg_press"} },
		"long free text":   func(p *workoutappv1.Profile) { p.FreeText = strings.Repeat("x", 501) },
	}
	for name, change := range cases {
		p := synthetic()
		change(p)
		if _, err := save(s, p); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: SaveProfile = %v, want InvalidArgument", name, err)
		}
	}
	if _, err := s.SaveProfile(signedIn, connect.NewRequest(&workoutappv1.SaveProfileRequest{})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no profile: SaveProfile = %v, want InvalidArgument", err)
	}
	if res, _ := s.GetProfile(signedIn, connect.NewRequest(&workoutappv1.GetProfileRequest{})); res.Msg.GetProfile() != nil {
		t.Fatal("a refused save stored a profile")
	}
}

type brokenStore struct{}

func (brokenStore) Get(context.Context, string) (profile.Profile, bool, error) {
	return profile.Profile{}, false, errors.New("rpc error: projects/p/databases/(default)/documents/users/uid-a")
}

func (brokenStore) Save(context.Context, string, profile.Profile) error {
	return errors.New("rpc error: projects/p/databases/(default)/documents/users/uid-a")
}

// TestStoreErrorIsHidden gives the caller a fixed text for a store
// error, so no path or uid leaves the server.
func TestStoreErrorIsHidden(t *testing.T) {
	s := New(brokenStore{})
	_, err := s.GetProfile(signedIn, connect.NewRequest(&workoutappv1.GetProfileRequest{}))
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users") {
		t.Fatalf("GetProfile = %v, want Internal with no path", err)
	}
	_, err = save(s, synthetic())
	if connect.CodeOf(err) != connect.CodeInternal || strings.Contains(err.Error(), "users") {
		t.Fatalf("SaveProfile = %v, want Internal with no path", err)
	}
}

// TestErrorHoldsNoText sends a synthetic health text in each field of a
// fixed list. Each error names the field alone, and never the text
// (D-80).
func TestErrorHoldsNoText(t *testing.T) {
	const text = "Synthetic note: knee pain after surgery"
	s := New(profile.NewMemory())
	cases := map[string]func(*workoutappv1.Profile){
		"experience":    func(p *workoutappv1.Profile) { p.Experience = text },
		"goal template": func(p *workoutappv1.Profile) { p.GoalTemplate = text },
		"muscle group":  func(p *workoutappv1.Profile) { p.MuscleGroups = []string{"chest", text} },
		"injured area":  func(p *workoutappv1.Profile) { p.InjuredAreas = []string{text} },
		"cardio":        func(p *workoutappv1.Profile) { p.CardioExerciseIds = []string{text} },
	}
	for name, change := range cases {
		p := synthetic()
		change(p)
		_, err := save(s, p)
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: SaveProfile = %v, want InvalidArgument", name, err)
			continue
		}
		if strings.Contains(err.Error(), "knee pain") {
			t.Errorf("%s: the error %q holds the text", name, err)
		}
	}
}
