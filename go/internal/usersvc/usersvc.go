// Package usersvc serves workoutapp.v1.UserService.
package usersvc

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"

	workoutappv1 "github.com/nkramber/workout-app/go/gen/workoutapp/v1"
	"github.com/nkramber/workout-app/go/gen/workoutapp/v1/workoutappv1connect"
	"github.com/nkramber/workout-app/go/internal/auth"
)

// Confirmation is the text that the owner types before a deletion of the
// history (D-314). The phone and the server check it.
const Confirmation = "Delete all data"

// History holds the stores of the history of a user (D-315). The
// profile, the inventory, the exclusions, the allowlist, and the monthly
// AI spend are not in it, so a deletion keeps them. Errors can be nil,
// for an API with no record of failed attempts.
type History struct {
	Workouts interface {
		DeleteAll(ctx context.Context, uid string) (int, error)
	}
	Plans interface {
		Delete(ctx context.Context, uid string) error
	}
	Errors interface {
		DeleteUser(ctx context.Context, uid string) (int, error)
	}
	Log *slog.Logger
}

// Server implements workoutappv1connect.UserServiceHandler. A Server
// with no History refuses DeleteHistory.
type Server struct {
	History *History
}

var _ workoutappv1connect.UserServiceHandler = Server{}

var errNoToken = connect.NewError(connect.CodeUnauthenticated, errors.New("a bearer token is required"))

// GetMe returns the uid that the auth interceptor stored. A call with no
// uid never reaches this point, but the check keeps the service safe
// without the interceptor too.
func (Server) GetMe(ctx context.Context, _ *connect.Request[workoutappv1.GetMeRequest]) (*connect.Response[workoutappv1.GetMeResponse], error) {
	uid := auth.UserID(ctx)
	if uid == "" {
		return nil, errNoToken
	}
	return connect.NewResponse(&workoutappv1.GetMeResponse{Uid: uid}), nil
}

// DeleteHistory deletes the workouts, the plan, and the AI error records
// of the caller (D-314, D-315). Each step deletes what an earlier failed
// call left, so the phone can call again. A failure gives UNAVAILABLE
// with no path, and the log line holds ids alone (D-80).
func (s Server) DeleteHistory(ctx context.Context, req *connect.Request[workoutappv1.DeleteHistoryRequest]) (*connect.Response[workoutappv1.DeleteHistoryResponse], error) {
	uid := auth.UserID(ctx)
	if uid == "" {
		return nil, errNoToken
	}
	if req.Msg.GetConfirmation() != Confirmation {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("delete history: the confirmation is not the text of D-314"))
	}
	h := s.History
	if h == nil || h.Workouts == nil || h.Plans == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, errors.New("delete history: the API has no history stores"))
	}
	fail := func(step string, err error) error {
		if h.Log != nil {
			h.Log.Warn("delete history failed", "uid", uid, "step", step, "err", err.Error())
		}
		return connect.NewError(connect.CodeUnavailable, errors.New("delete history: the "+step+" stayed, try again"))
	}
	workouts, err := h.Workouts.DeleteAll(ctx, uid)
	if err != nil {
		return nil, fail("workouts", err)
	}
	if err := h.Plans.Delete(ctx, uid); err != nil {
		return nil, fail("plan", err)
	}
	records := 0
	if h.Errors != nil {
		if records, err = h.Errors.DeleteUser(ctx, uid); err != nil {
			return nil, fail("error records", err)
		}
	}
	if h.Log != nil {
		h.Log.Info("history deleted", "uid", uid, "workouts", workouts, "error_records", records)
	}
	return connect.NewResponse(&workoutappv1.DeleteHistoryResponse{DeletedWorkouts: int32(min(workouts, 1<<31-1))}), nil
}
