package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/nkramber/workout-app/go/internal/domain"
	"github.com/nkramber/workout-app/go/internal/policy"
)

// Call is one request to a provider: the role, the instructions, the
// input JSON, and the strict schema of the output.
type Call struct {
	Role         Role
	Instructions string
	Input        []byte
	SchemaName   string
	Schema       []byte
}

// Reply is the answer of a provider. Refusal tells that the model
// refused the request. Incomplete tells that the output stopped early,
// for example at the output limit.
type Reply struct {
	Text       string
	Usage      Usage
	Refusal    bool
	Incomplete bool
}

// Provider sends one call and gives the reply. It makes one attempt,
// with no retry: a failure gives the rules fallback (D-23). An error
// holds no key, no prompt, and no output text (D-80).
type Provider interface {
	Send(ctx context.Context, c Call) (Reply, error)
}

// Fake is the provider of the tests (D-24). It makes no network call.
// Reply gives the answer of each call. When Hang is true, Send waits
// until the context ends, as a call that times out. Delay makes each
// call wait before its reply, so a browser test can see the progress of
// a request (D-241). Fake is safe for use by more than one goroutine.
type Fake struct {
	Reply func(Call) (Reply, error)
	Hang  bool
	Delay time.Duration

	mu    sync.Mutex
	calls []Call
}

// Send records the call and gives the reply of the fake.
func (f *Fake) Send(ctx context.Context, c Call) (Reply, error) {
	f.mu.Lock()
	f.calls = append(f.calls, c)
	f.mu.Unlock()
	if f.Hang {
		<-ctx.Done()
		return Reply{}, ctx.Err()
	}
	if f.Delay > 0 {
		select {
		case <-ctx.Done():
			return Reply{}, ctx.Err()
		case <-time.After(f.Delay):
		}
	}
	if f.Reply == nil {
		return EchoReply(c)
	}
	return f.Reply(c)
}

// Calls gives a copy of each call that the fake got.
func (f *Fake) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Call(nil), f.calls...)
}

// EchoReply gives a valid output for a call: each session holds the
// first MaxSessionExercises exercises of the input at their policy
// targets, with the default warm-up and cool-down. When the rotation
// rules apply (D-328), each session holds the exercises of its class of
// echoSplit alone. When the input has a
// cardio exercise, each session gets EchoCardioMinutes of the first one.
// The plan holds the EchoGuidance items, so a test sees each part of a
// plan (D-44, D-241). The usage counts 4 bytes as one token. It is the
// default reply of the fake.
// The cardio time and the guidance items of EchoReply. The cardio time
// is the least time of D-255.
const EchoCardioMinutes = MinCardioMinutes

var EchoGuidance = []GuidanceID{"mobility.hips", "recovery.rest_day"}

func EchoReply(c Call) (Reply, error) {
	var in wireInput
	if err := json.Unmarshal(c.Input, &in); err != nil {
		return Reply{}, errors.New("fake: the input is not the JSON of the layer")
	}
	if c.SchemaName == ReasonSchemaName {
		return echoReasons(c, in), nil
	}
	type exercise struct {
		ID          string        `json:"exercise_id"`
		Rest        int           `json:"rest_seconds"`
		Calibration []wireCalSet  `json:"calibration_sets"`
		Working     []wireWorkSet `json:"working_sets"`
		Reason      string        `json:"reason"`
	}
	type cardio struct {
		ID      string `json:"exercise_id"`
		Minutes int    `json:"minutes"`
	}
	type session struct {
		WarmUp    GuidanceID `json:"warm_up_id"`
		Exercises []exercise `json:"exercises"`
		Cardio    cardio     `json:"cardio"`
		CoolDown  GuidanceID `json:"cool_down_id"`
	}
	out := struct {
		Summary  string       `json:"summary"`
		Sessions []session    `json:"sessions"`
		Guidance []GuidanceID `json:"guidance_ids"`
	}{Summary: "A plan at the targets of the rules.", Guidance: append([]GuidanceID(nil), EchoGuidance...)}
	split := echoSplit(in)
	for i := range in.Sessions {
		s := session{WarmUp: DefaultWarmUp, CoolDown: DefaultCoolDown, Exercises: []exercise{}}
		if len(in.Cardio) > 0 {
			s.Cardio = cardio{in.Cardio[0], EchoCardioMinutes}
		}
		for _, e := range split[i] {
			s.Exercises = append(s.Exercises, exercise{e.ID, e.Target.Rest, e.Target.Calibration, e.Target.Working,
				"The target follows your last logged sets."})
		}
		out.Sessions = append(out.Sessions, s)
	}
	return echoText(c, encode(out)), nil
}

// echoSplit gives the exercises of each session of EchoReply, in the
// order of the input. With no rotation, each session holds the first
// MaxSessionExercises exercises. With the rotation, the units of the
// policy go to the classes of sessions in turn, and an exercise with no
// group goes to each session. A session first takes each exercise that
// adds a selected group, so that the cover rule holds.
func echoSplit(in wireInput) [][]wireExerciseIn {
	out := make([][]wireExerciseIn, in.Sessions)
	if in.Rotation == nil || !*in.Rotation {
		for i := range out {
			out[i] = in.Exercises[:min(len(in.Exercises), MaxSessionExercises)]
		}
		return out
	}
	r := policy.Rotation{Sessions: in.Sessions, Groups: map[domain.ExerciseID][]domain.MuscleGroup{}}
	for _, e := range in.Exercises {
		id := domain.ExerciseID(e.ID)
		r.Order = append(r.Order, id)
		for _, g := range e.Groups {
			r.Groups[id] = append(r.Groups[id], domain.MuscleGroup(g))
		}
	}
	if in.Profile != nil {
		for _, g := range in.Profile.MuscleGroups {
			r.Selected = append(r.Selected, domain.MuscleGroup(g))
		}
	}
	class := map[domain.ExerciseID]int{}
	for j, u := range r.Units() {
		for _, id := range u {
			class[id] = j % r.Classes()
		}
	}
	required := r.Required()
	for i := range out {
		var mine []int
		for k, e := range in.Exercises {
			if c, ok := class[domain.ExerciseID(e.ID)]; !ok || c == i%r.Classes() {
				mine = append(mine, k)
			}
		}
		var picked []int
		covered := map[string]bool{}
		for _, k := range mine {
			adds := slices.ContainsFunc(in.Exercises[k].Groups, func(g string) bool {
				return !covered[g] && slices.Contains(required, domain.MuscleGroup(g))
			})
			if adds && len(picked) < MaxSessionExercises {
				picked = append(picked, k)
				for _, g := range in.Exercises[k].Groups {
					covered[g] = true
				}
			}
		}
		for _, k := range mine {
			if len(picked) < MaxSessionExercises && !slices.Contains(picked, k) {
				picked = append(picked, k)
			}
		}
		slices.Sort(picked)
		for _, k := range picked {
			out[i] = append(out[i], in.Exercises[k])
		}
	}
	return out
}

func echoText(c Call, out []byte) Reply {
	text := string(out)
	return Reply{Text: text, Usage: Usage{
		InputTokens:  int64(len(c.Instructions)+len(c.Input)) / 4,
		OutputTokens: int64(len(text)) / 4,
	}}
}

// EchoReason gives the reason of the fake reviser for the first working
// set of the last session: "Set 1: 12 reps at 25 lb, 1 in reserve." It
// names that set, so the check of D-288 accepts it.
func EchoReason(reps int, weightLb float64, rir int) string {
	return fmt.Sprintf("Set 1: %d reps at %s lb, %d in reserve.", reps, strconv.FormatFloat(weightLb, 'f', -1, 64), rir)
}

// echoReasons gives a valid reviser output: for each exercise whose last
// session has a working set, the reason of EchoReason. An exercise with
// no such set gets no reason, so the reason of the rules shows.
func echoReasons(c Call, in wireInput) Reply {
	type set struct {
		Kind   string `json:"kind"`
		Number int    `json:"number"`
	}
	type reason struct {
		ID     string `json:"exercise_id"`
		Sets   []set  `json:"logged_sets"`
		Reason string `json:"reason"`
	}
	out := struct {
		Reasons []reason `json:"reasons"`
	}{Reasons: []reason{}}
	for _, e := range in.Exercises {
		if len(e.History) == 0 {
			continue
		}
		for _, s := range e.History[len(e.History)-1].Sets {
			if s.Kind == string(domain.SetWorking) && s.Number == 1 {
				out.Reasons = append(out.Reasons, reason{e.ID, []set{{s.Kind, 1}}, EchoReason(s.Reps, s.Weight, s.RIR)})
				break
			}
		}
	}
	return echoText(c, encode(out))
}
