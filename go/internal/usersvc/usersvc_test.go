package usersvc

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
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
