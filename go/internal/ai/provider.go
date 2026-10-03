package ai

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
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
// targets, with the default warm-up and cool-down. When the input has a
// cardio exercise, each session gets EchoCardioMinutes of the first one.
// The plan holds the EchoGuidance items, so a test sees each part of a
// plan (D-44, D-241). The usage counts 4 bytes as one token. It is the
// default reply of the fake.
// The cardio time and the guidance items of EchoReply.
const EchoCardioMinutes = 10

var EchoGuidance = []GuidanceID{"mobility.hips", "recovery.rest_day"}

func EchoReply(c Call) (Reply, error) {
	var in wireInput
	if err := json.Unmarshal(c.Input, &in); err != nil {
		return Reply{}, errors.New("fake: the input is not the JSON of the layer")
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
	for range in.Sessions {
		s := session{WarmUp: DefaultWarmUp, CoolDown: DefaultCoolDown, Exercises: []exercise{}}
		if len(in.Cardio) > 0 {
			s.Cardio = cardio{in.Cardio[0], EchoCardioMinutes}
		}
		for _, e := range in.Exercises[:min(len(in.Exercises), MaxSessionExercises)] {
			s.Exercises = append(s.Exercises, exercise{e.ID, e.Target.Rest, e.Target.Calibration, e.Target.Working,
				"The target follows your last logged sets."})
		}
		out.Sessions = append(out.Sessions, s)
	}
	text := string(encode(out))
	return Reply{Text: text, Usage: Usage{
		InputTokens:  int64(len(c.Instructions)+len(c.Input)) / 4,
		OutputTokens: int64(len(text)) / 4,
	}}, nil
}
