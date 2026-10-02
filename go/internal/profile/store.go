package profile

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

// The Firestore path of the one profile of a user is
// users/{uid}/profile/active, in the form of the inventory path (D-197).
// One document holds the whole profile, and each save replaces it.
const (
	UsersCollection = "users"
	Collection      = "profile"
	ActiveDoc       = "active"
)

// Store reads and saves the profile of a user. Get gives false when the
// user saved no profile. Save replaces the whole profile as one atomic
// write.
type Store interface {
	Get(ctx context.Context, uid string) (Profile, bool, error)
	Save(ctx context.Context, uid string, p Profile) error
}

var errUID = errors.New("profile: a uid of 1 or more characters with no slash is required")

func checkUID(uid string) error {
	if uid == "" || strings.Contains(uid, "/") {
		return errUID
	}
	return nil
}

// Memory is a Store in memory, for tests.
type Memory struct {
	mu   sync.Mutex
	data map[string]Profile
}

// NewMemory gives an empty Memory store.
func NewMemory() *Memory { return &Memory{data: map[string]Profile{}} }

// Get gives the profile of the uid.
func (s *Memory) Get(_ context.Context, uid string) (Profile, bool, error) {
	if err := checkUID(uid); err != nil {
		return Profile{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.data[uid]
	return p.clone(), ok, nil
}

// Save stores a copy of p.
func (s *Memory) Save(_ context.Context, uid string, p Profile) error {
	if err := checkUID(uid); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[uid] = p.clone()
	return nil
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

// Get reads the document of the uid.
func (s *Firestore) Get(ctx context.Context, uid string) (Profile, bool, error) {
	if err := checkUID(uid); err != nil {
		return Profile{}, false, err
	}
	snap, err := s.ref(uid).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return Profile{}, false, nil
	}
	if err != nil {
		return Profile{}, false, err
	}
	var d document
	if err := snap.DataTo(&d); err != nil {
		return Profile{}, false, err
	}
	return d.profile(), true, nil
}

// Save writes the whole document. One write of one document is atomic,
// so a save needs no transaction.
func (s *Firestore) Save(ctx context.Context, uid string, p Profile) error {
	if err := checkUID(uid); err != nil {
		return err
	}
	_, err := s.ref(uid).Set(ctx, encode(p))
	return err
}

// The stored document. Each value is an id of a fixed list or of the
// catalog, a whole number, or a text.
type document struct {
	Experience   string   `firestore:"experience"`
	Template     string   `firestore:"goal_template"`
	Groups       []string `firestore:"muscle_groups"`
	FreeText     string   `firestore:"free_text"`
	InjuredAreas []string `firestore:"injured_areas"`
	InjuryText   string   `firestore:"injury_text"`
	AgeYears     int64    `firestore:"age_years"`
	HeightIn     int64    `firestore:"height_in"`
	WeightLb     int64    `firestore:"weight_lb"`
	Cardio       []string `firestore:"cardio_exercises"`
	TrainingDays int64    `firestore:"training_days"`
}

func (d document) profile() Profile {
	return Profile{
		Experience:   Experience(d.Experience),
		Template:     domain.TemplateID(d.Template),
		Groups:       convert[domain.MuscleGroup](d.Groups),
		FreeText:     d.FreeText,
		InjuredAreas: convert[domain.Area](d.InjuredAreas),
		InjuryText:   d.InjuryText,
		AgeYears:     int(d.AgeYears),
		HeightIn:     int(d.HeightIn),
		WeightLb:     int(d.WeightLb),
		Cardio:       convert[domain.ExerciseID](d.Cardio),
		TrainingDays: int(d.TrainingDays),
	}
}

func encode(p Profile) document {
	return document{
		Experience:   string(p.Experience),
		Template:     string(p.Template),
		Groups:       convert[string](p.Groups),
		FreeText:     p.FreeText,
		InjuredAreas: convert[string](p.InjuredAreas),
		InjuryText:   p.InjuryText,
		AgeYears:     int64(p.AgeYears),
		HeightIn:     int64(p.HeightIn),
		WeightLb:     int64(p.WeightLb),
		Cardio:       convert[string](p.Cardio),
		TrainingDays: int64(p.TrainingDays),
	}
}

// convert gives a new list of the values as type To. An empty list
// gives an empty list, not nil, so Firestore stores an empty array.
func convert[To, From ~string](in []From) []To {
	out := make([]To, 0, len(in))
	for _, v := range in {
		out = append(out, To(v))
	}
	return out
}
