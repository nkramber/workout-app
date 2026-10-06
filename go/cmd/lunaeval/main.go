// Command lunaeval is the Phase 3 evaluation of the Luna role layer
// (work areas 3.2 and 3.3). It sends the synthetic profiles of
// profiles.json through the planner, and the scenarios A to F of
// section 5 of the high-level roadmap through the reviser. The policy
// of "go/internal/policy" then decides each exercise (D-23). The
// command writes a JSON report with the schema pass rate, the policy
// refusals, the grade of each scenario, and the cost.
//
// The -effort flag replaces the reasoning effort of each role, so two
// runs can compare two efforts with the same calls.
//
// With no -live flag, the command uses the fake provider, and it costs
// nothing. With -live, it calls OpenAI, and each call costs money. A
// live run needs the approval of the owner, with its cap, at run time
// (D-25). The key comes from the environment variable OPENAI_API_KEY
// alone, and the command never writes it.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/nkramber/workout-app/go/internal/ai"
)

// EnvKey is the environment variable that holds the OpenAI key.
const EnvKey = "OPENAI_API_KEY"

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "lunaeval:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout io.Writer) error {
	fs := flag.NewFlagSet("lunaeval", flag.ContinueOnError)
	live := fs.Bool("live", false, "call OpenAI: each call costs money, and the owner approves the run and its cap (D-25)")
	capUSD := fs.String("cap", "", "the cap of the run in US dollars, such as 2 (required)")
	repeats := fs.Int("repeats", 5, "the calls of each scenario")
	workers := fs.Int("workers", 4, "the calls at the same time")
	out := fs.String("out", "", "the path of the JSON report (required)")
	effort := fs.String("effort", "", "the reasoning effort of each call, one of "+strings.Join(ai.Efforts, ", ")+" (default: the effort of the role)")
	planner := fs.Bool("planner", true, "send the profiles through the planner: false runs the reviser scenarios alone")
	reviser := fs.Bool("reviser", true, "run the reviser scenarios: false runs the planner alone")
	sessions := fs.Int("sessions", PlannerSessions, "the sessions of each planner call, 2 to 4 (D-211)")
	plannerRepeats := fs.Int("planner-repeats", 1, "the planner calls of each profile")
	groups := fs.Bool("groups", false, "give each profile the template \"General fitness\" with each muscle group, so that each rotation rule applies (D-328)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *out == "" || *capUSD == "" {
		return errors.New("-out and -cap are required")
	}
	if *repeats < 1 || *workers < 1 || *plannerRepeats < 1 {
		return errors.New("-repeats, -planner-repeats, and -workers must be 1 or more")
	}
	if *sessions < 2 || *sessions > 4 {
		return errors.New("-sessions must be 2 to 4")
	}
	if *effort != "" && !ai.ValidEffort(*effort) {
		return fmt.Errorf("-effort %q: want one of %s", *effort, strings.Join(ai.Efforts, ", "))
	}
	// The cap is the cap of the user and of the project of the run.
	caps, err := ai.CapsFromEnv(func(string) string { return *capUSD })
	if err != nil {
		return err
	}
	var provider ai.Provider = &ai.Fake{}
	name := "fake"
	if *live {
		p, err := ai.NewOpenAI(getenv(EnvKey), "", nil)
		if err != nil {
			return fmt.Errorf("-live needs %s: %w", EnvKey, err)
		}
		provider, name = p, "openai"
	}
	var profiles []Profile
	if *planner {
		if profiles, err = Profiles(); err != nil {
			return err
		}
	}
	client := &ai.Client{Provider: provider, Cap: ai.NewMemoryCap(caps), Effort: *effort}
	var scenarios []Scenario
	if *reviser {
		scenarios = Scenarios()
	}
	f := Plans{Sessions: *sessions, Repeats: *plannerRepeats, Groups: *groups}
	rep, err := Run(ctx, client, profiles, f, scenarios, *repeats, *workers)
	if err != nil {
		return err
	}
	rep.Provider, rep.Cap = name, caps.Project.String()
	data, err := json.MarshalIndent(rep, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(*out, append(data, '\n'), 0o644); err != nil {
		return err
	}
	summary(stdout, rep)
	return nil
}

// summary writes the counts of a report. It holds ids and numbers
// alone (D-80).
func summary(w io.Writer, r Report) {
	t := r.Totals
	fmt.Fprintf(w, "provider %s, model %s, effort %s, cap %s\n", r.Provider, r.Model, r.Effort, r.Cap)
	fmt.Fprintf(w, "calls %d, schema pass %d, cost %s, unknown cost %d, max %.1f s\n", t.Calls, t.SchemaPass, t.Cost, t.CostUnknown, t.MaxSeconds)
	var st []string
	for s, n := range t.ByStatus {
		st = append(st, fmt.Sprintf("%s %d", s, n))
	}
	sort.Strings(st)
	fmt.Fprintf(w, "status: %v\n", st)
	fmt.Fprintf(w, "decisions %d, proposals %d, accepted %d, refused %d, no proposal %d, not planned %d\n",
		t.Decisions, t.Proposals, t.Accepted, t.Refused, t.NoProposal, t.NotPlanned)
	fmt.Fprintf(w, "refusals by rule: %v, filtered texts %d, cardio items %d\n", t.ByRule, t.Filtered, t.Cardio)
	fmt.Fprintf(w, "revisions %d, luna reasons %d, rules reasons by cause %v\n", t.Revisions, t.LunaReasons, t.ReasonCauses)
	fmt.Fprintf(w, "rotation applies %d, rotation refused %d\n", t.RotationApplies, t.RotationRefused)
	for _, s := range r.Scenarios {
		fmt.Fprintf(w, "scenario %s: pass %v, safe %d of %d, reasons accepted %d, refused %d %v, no reason read %d, jumps refused %d of %d, calls %d, cost %s, max %.1f s\n",
			s.ID, s.Pass(), s.Safe, s.Cases, s.LunaReasons, s.Refused, s.ByCause, s.NoReason, s.JumpsOK, s.Jumps, s.Calls, s.Cost, s.MaxSeconds)
	}
}
