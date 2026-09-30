// Package usersvc serves workoutapp.v1.UserService.
package usersvc

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
)

// Server implements workoutappv1connect.UserServiceHandler.
type Server struct{}

var _ workoutappv1connect.UserServiceHandler = Server{}

// GetMe returns the uid that the auth interceptor stored. A call with no
// uid never reaches this point, but the check keeps the service safe
// without the interceptor too.
func (Server) GetMe(ctx context.Context, _ *connect.Request[workoutappv1.GetMeRequest]) (*connect.Response[workoutappv1.GetMeResponse], error) {
	uid := auth.UserID(ctx)
	if uid == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))
	}
	return connect.NewResponse(&workoutappv1.GetMeResponse{Uid: uid}), nil
}
