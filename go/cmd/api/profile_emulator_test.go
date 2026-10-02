//go:build emulator

// The acceptance story of PR-23 over the Firebase emulators. The test
// saves and reads a profile through the API, and the server refuses a
// field outside its bound. Then it reads the stored document, and the
// planner input of profile.ForPlan.
package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/allowlist"
	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/profile"
)

func TestProfileAcceptanceStory(t *testing.T) {
	authHost := requireEmulators(t)
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	uid, token := signUp(t, authHost, fmt.Sprintf("profile-%d@example.test", suffix))
	otherUID, otherToken := signUp(t, authHost, fmt.Sprintf("profile-other-%d@example.test", suffix))
	_, outsideToken := signUp(t, authHost, fmt.Sprintf("profile-outside-%d@example.test", suffix))

	fs, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = fs.Close() }()
	for _, id := range []string{uid, otherUID} {
		if _, err := fs.Collection(allowlist.Collection).Doc(id).Set(ctx, map[string]any{}); err != nil {
			t.Fatal(err)
		}
	}
	client := workoutappv1connect.NewProfileServiceClient(http.DefaultClient, startAPI(t))

	// The fixed lists come from the server.
	opts, err := client.GetProfileOptions(ctx, signed(token, &workoutappv1.GetProfileOptionsRequest{}))
	if err != nil || len(opts.Msg.GetInjuryAreas()) != len(domain.Areas()) || len(opts.Msg.GetMuscleGroups()) != len(domain.MuscleGroups()) {
		t.Fatalf("GetProfileOptions = %v, %v", opts, err)
	}

	// A new user has no profile.
	got, err := client.GetProfile(ctx, signed(token, &workoutappv1.GetProfileRequest{}))
	if err != nil || got.Msg.GetProfile() != nil {
		t.Fatalf("GetProfile before a save = %v, %v, want no profile", got, err)
	}

	// Save a synthetic profile, and read it again.
	in := &workoutappv1.Profile{
		Experience:        "intermediate",
		GoalTemplate:      "general_fitness",
		MuscleGroups:      []string{"chest", "back", "quadriceps"},
		FreeText:          "More upper body.",
		InjuredAreas:      []string{"knee"},
		InjuryText:        "Old knee strain.",
		AgeYears:          32,
		HeightIn:          70,
		WeightLb:          180,
		CardioExerciseIds: []string{"upright_bike", "rowing_machine"},
		TrainingDays:      3,
	}
	saved, err := client.SaveProfile(ctx, signed(token, &workoutappv1.SaveProfileRequest{Profile: in}))
	if err != nil {
		t.Fatal(err)
	}
	got, err = client.GetProfile(ctx, signed(token, &workoutappv1.GetProfileRequest{}))
	if err != nil || got.Msg.GetProfile().String() != saved.Msg.GetProfile().String() || got.Msg.GetProfile().GetWeightLb() != 180 {
		t.Fatalf("GetProfile = %v, %v, want %v", got, err, saved.Msg.GetProfile())
	}

	// The server refuses a field outside its bound, and keeps the stored
	// profile.
	for name, change := range map[string]func(*workoutappv1.Profile){
		"age 17":          func(p *workoutappv1.Profile) { p.AgeYears = 17 },
		"weight 501":      func(p *workoutappv1.Profile) { p.WeightLb = 501 },
		"height 47":       func(p *workoutappv1.Profile) { p.HeightIn = 47 },
		"1 training day":  func(p *workoutappv1.Profile) { p.TrainingDays = 1 },
		"area off a list": func(p *workoutappv1.Profile) { p.InjuredAreas = []string{"neck"} },
		"beginner":        func(p *workoutappv1.Profile) { p.Experience = "beginner" },
	} {
		bad := proto.Clone(saved.Msg.GetProfile()).(*workoutappv1.Profile)
		change(bad)
		if _, err := client.SaveProfile(ctx, signed(token, &workoutappv1.SaveProfileRequest{Profile: bad})); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: SaveProfile = %v, want InvalidArgument", name, err)
		}
	}
	got, err = client.GetProfile(ctx, signed(token, &workoutappv1.GetProfileRequest{}))
	if err != nil || got.Msg.GetProfile().String() != saved.Msg.GetProfile().String() {
		t.Fatalf("after the refused saves: %v, %v", got, err)
	}

	// Each uid has its own profile, and a uid off the allowlist gets no
	// call.
	if other, err := client.GetProfile(ctx, signed(otherToken, &workoutappv1.GetProfileRequest{})); err != nil || other.Msg.GetProfile() != nil {
		t.Fatalf("the other uid reads %v, %v", other, err)
	}
	if _, err := client.GetProfile(ctx, signed(outsideToken, &workoutappv1.GetProfileRequest{})); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("a uid off the allowlist = %v, want PermissionDenied", err)
	}

	// The stored document holds the profile, and the planner input holds
	// the inputs of D-209 alone, with no exercise of the injured knee.
	stored, ok, err := profile.FromFirestore(fs).Get(ctx, uid)
	if err != nil || !ok || stored.WeightLb != 180 || !slices.Equal(stored.InjuredAreas, []domain.Area{domain.AreaKnee}) {
		t.Fatalf("stored = %+v, %v, %v", stored, ok, err)
	}
	tables := domain.DefaultBodyTables()
	plan := profile.ForPlan(stored, domain.DefaultCatalog(), tables)
	for _, id := range append(slices.Clone(plan.Exercises), plan.Cardio...) {
		if slices.Contains(tables.Areas[id], domain.AreaKnee) {
			t.Errorf("the planner input holds %q, which loads the knee", id)
		}
	}
	if plan.FreeText != "More upper body." || plan.Sessions != 3 || len(plan.Exercises) == 0 {
		t.Fatalf("planner input = %+v", plan)
	}
}
