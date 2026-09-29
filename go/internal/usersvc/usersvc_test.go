package usersvc

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	gymroutev1 "github.com/nkramber/workout-app/go/gen/gymroute/v1"
	"github.com/nkramber/workout-app/go/internal/auth"
)

func TestGetMe(t *testing.T) {
	req := connect.NewRequest(&gymroutev1.GetMeRequest{})
	res, err := Server{}.GetMe(auth.WithUserID(context.Background(), "uid-a"), req)
	if err != nil || res.Msg.GetUid() != "uid-a" {
		t.Fatalf("GetMe = %v, %v, want uid-a", res, err)
	}
	if _, err := (Server{}).GetMe(context.Background(), req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("GetMe with no uid = %v, want Unauthenticated", err)
	}
}
