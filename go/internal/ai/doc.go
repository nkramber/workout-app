// Package ai holds the Luna role layer (D-22, D-24). The planner and
// the reviser call OpenAI Luna at medium effort with a strict JSON
// schema. A call site names a role, and the role gives the model id, so
// no model id appears at a call site. The fake provider stands
// in for Luna in each test, and no test calls OpenAI.
//
// Each call has a cost record, and a cap hook reserves the worst-case
// cost before the call. The hook refuses a call over the cap, and the
// cap values come from the configuration (D-25).
//
// Luna writes one plan summary and one short reason for each exercise
// (D-182). A versioned filter of blocked claims reads each such text,
// and a template text replaces a text that it blocks (D-183). Session
// titles come from a template. Warm-up, cool-down, mobility, and
// recovery texts come from the versioned guidance catalog, and Luna
// selects them by id (D-152). The prompt keeps each text inside the
// fitness boundary of D-36 and the dated copy of the usage policies of
// D-93.
//
// The layer gives a proposal for each exercise. The policy of
// "go/internal/policy" checks it through policy.Decide before the owner
// sees a target (D-23). A call that gives no valid proposal gives a
// proposal with no target, and the policy gives the rules fallback.
//
// A cost record and an error hold ids and numbers alone, so each can go
// into a log (D-80). A plan holds the text of Luna and the targets of
// the owner, and it never goes into a log.
package ai
