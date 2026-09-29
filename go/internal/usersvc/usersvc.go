// Package usersvc serves gymroute.v1.UserService.
package usersvc

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	gymroutev1 "github.com/nkramber/workout-app/go/gen/gymroute/v1"
	"github.com/nkramber/workout-app/go/gen/gymroute/v1/gymroutev1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
)

// Server implements gymroutev1connect.UserServiceHandler.
type Server struct{}

var _ gymroutev1connect.UserServiceHandler = Server{}

// GetMe returns the uid that the auth interceptor stored. A call with no
// uid never reaches this point, but the check keeps the service safe
// without the interceptor too.
func (Server) GetMe(ctx context.Context, _ *connect.Request[gymroutev1.GetMeRequest]) (*connect.Response[gymroutev1.GetMeResponse], error) {
	uid := auth.UserID(ctx)
	if uid == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))
	}
	return connect.NewResponse(&gymroutev1.GetMeResponse{Uid: uid}), nil
}
