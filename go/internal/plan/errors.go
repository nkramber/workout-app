package plan

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/nkramber/workout-app/go/internal/ai"
)

// ErrorsCollection is the top-level collection of the error records of
// the failed attempts (D-236). A Firestore TTL policy on the field
// expire_at deletes each record after ErrorRetention.
const ErrorsCollection = "aiErrors"

// ErrorRetention is the time that an error record stays (D-236).
const ErrorRetention = 90 * 24 * time.Hour

// ErrorRecord is the record of one failed attempt, for a later analysis
// of the prompt (D-236). Output is the output text of Luna, which can
// repeat a text of the owner. So the record never goes into a log, a
// metric, or an error report (D-80), and the rules refuse each client.
type ErrorRecord struct {
	User          string
	Time          time.Time
	ExpireAt      time.Time
	Request       Kind
	Attempt       int
	MaxAttempts   int
	Status        ai.Status
	Cause         string
	Model         string
	Effort        string
	PromptVersion string
	PromptHash    string
	SchemaName    string
	Cost          ai.CostRecord
	Output        string
}

// Kind names the request of an attempt.
type Kind string

const (
	KindPlan    Kind = "plan"
	KindExclude Kind = "exclude"
)

// ErrorLog adds error records.
type ErrorLog interface {
	Add(ctx context.Context, r ErrorRecord) error
}

// MemoryErrors is an ErrorLog in memory, for tests.
type MemoryErrors struct {
	mu      sync.Mutex
	records []ErrorRecord
}

// Add keeps a copy of the record.
func (m *MemoryErrors) Add(_ context.Context, r ErrorRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, r)
	return nil
}

// Records gives a copy of each record, in the order of Add.
func (m *MemoryErrors) Records() []ErrorRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ErrorRecord(nil), m.records...)
}

// FirestoreErrors is the production ErrorLog. Each record is a new
// document with an id from Firestore.
type FirestoreErrors struct {
	client *firestore.Client
}

// ErrorsFromFirestore gives the ErrorLog of the client.
func ErrorsFromFirestore(client *firestore.Client) *FirestoreErrors {
	return &FirestoreErrors{client: client}
}

// Add writes the record as a new document.
func (f *FirestoreErrors) Add(ctx context.Context, r ErrorRecord) error {
	_, _, err := f.client.Collection(ErrorsCollection).Add(ctx, encodeError(r))
	return err
}

type errorDoc struct {
	User          string    `firestore:"uid"`
	Time          time.Time `firestore:"time"`
	ExpireAt      time.Time `firestore:"expire_at"`
	Request       string    `firestore:"request"`
	Attempt       int64     `firestore:"attempt"`
	MaxAttempts   int64     `firestore:"max_attempts"`
	Status        string    `firestore:"status"`
	Cause         string    `firestore:"cause"`
	Model         string    `firestore:"model"`
	Effort        string    `firestore:"effort"`
	PromptVersion string    `firestore:"prompt_version"`
	PromptHash    string    `firestore:"prompt_hash"`
	SchemaName    string    `firestore:"schema"`
	InputTokens   int64     `firestore:"input_tokens"`
	OutputTokens  int64     `firestore:"output_tokens"`
	Reserved      int64     `firestore:"reserved_nano_usd"`
	Cost          int64     `firestore:"cost_nano_usd"`
	CostKnown     bool      `firestore:"cost_known"`
	Unsettled     bool      `firestore:"unsettled"`
	Output        string    `firestore:"output"`
}

func encodeError(r ErrorRecord) errorDoc {
	return errorDoc{
		User: r.User, Time: r.Time.UTC(), ExpireAt: r.ExpireAt.UTC(), Request: string(r.Request),
		Attempt: int64(r.Attempt), MaxAttempts: int64(r.MaxAttempts), Status: string(r.Status), Cause: r.Cause,
		Model: r.Model, Effort: r.Effort, PromptVersion: r.PromptVersion, PromptHash: r.PromptHash, SchemaName: r.SchemaName,
		InputTokens: r.Cost.Usage.InputTokens, OutputTokens: r.Cost.Usage.OutputTokens,
		Reserved: int64(r.Cost.Reserved), Cost: int64(r.Cost.Cost), CostKnown: r.Cost.Known, Unsettled: r.Cost.Unsettled,
		Output: r.Output,
	}
}
