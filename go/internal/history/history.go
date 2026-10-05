// Package history holds the fence of a deletion of the history (D-315).
// The document users/{uid}/history/deleted holds the generation of the
// history: the count of the deletions. Each deletion adds 1 to it first.
// A write of a workout or of a plan reads it in its transaction, and
// refuses a write of an older generation. So a sync or a plan request
// that started before a deletion never writes old history after it. The
// generation comes from the server alone, and no clock of a phone
// decides.
package history

import (
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The path of the fence: users/{uid}/history/deleted.
const (
	UsersCollection = "users"
	Collection      = "history"
	DeletedDoc      = "deleted"
)

// Fence is the stored generation and the time of the last deletion, on
// the clock of the server.
type Fence struct {
	Generation int64     `firestore:"generation"`
	At         time.Time `firestore:"deleted_at"`
}

// Ref gives the document of the fence of the uid.
func Ref(client *firestore.Client, uid string) *firestore.DocumentRef {
	return client.Collection(UsersCollection).Doc(uid).Collection(Collection).Doc(DeletedDoc)
}

// Read gives the fence of a read of Ref, or a zero fence for no
// document.
func Read(snap *firestore.DocumentSnapshot, err error) (Fence, error) {
	switch {
	case err == nil:
		var f Fence
		if err := snap.DataTo(&f); err != nil {
			return Fence{}, err
		}
		return f, nil
	case status.Code(err) == codes.NotFound:
		return Fence{}, nil
	default:
		return Fence{}, err
	}
}
