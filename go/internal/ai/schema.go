package ai

import (
	"encoding/json"
	"slices"
	"sync"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// SchemaName is the name and the version of the plan schema. The plan
// spike used version 1 ("docs/research/luna-plan-spike.md"). Version 2
// keeps its strict form, names the ids of the catalog of D-155, and
// replaces the free text of the titles, the warm-up, the cool-down, and
// the guidance with ids of the guidance catalog (D-152, D-182).
const SchemaName = "luna_plan_v2"

// Schema gives the strict JSON schema of the output of Luna. The form
// obeys the strict structured outputs of OpenAI: each property is
// required, and additionalProperties is false. An enum holds each id,
// so a valid output names no unknown exercise and no unknown guidance
// item. A cardio of exercise "" and 0 minutes means no cardio. The
// policy checks the values, not the schema (D-23).
func Schema() []byte { return schemaOnce() }

var schemaOnce = sync.OnceValue(func() []byte {
	var strength, cardio []string
	for _, e := range domain.DefaultCatalog().Exercises {
		if e.Kind == domain.KindCardio {
			cardio = append(cardio, string(e.ID))
		} else {
			strength = append(strength, string(e.ID))
		}
	}
	cardio = append(cardio, "")
	ids := func(kinds ...GuidanceKind) []string {
		var out []string
		for _, id := range guidanceIDs(kinds...) {
			out = append(out, string(id))
		}
		return out
	}
	set := func(rir bool) obj {
		p := obj{"reps": typ("integer"), "load_lb": typ("number")}
		if rir {
			p["rir_target"] = typ("integer")
		}
		return object(p)
	}
	exercise := object(obj{
		"exercise_id":      enum(strength),
		"rest_seconds":     typ("integer"),
		"calibration_sets": array(set(false)),
		"working_sets":     array(set(true)),
		"reason":           typ("string"),
	})
	session := object(obj{
		"warm_up_id":   enum(ids(KindWarmUp)),
		"exercises":    array(exercise),
		"cardio":       object(obj{"exercise_id": enum(cardio), "minutes": typ("integer")}),
		"cool_down_id": enum(ids(KindCoolDown)),
	})
	plan := object(obj{
		"summary":      typ("string"),
		"sessions":     array(session),
		"guidance_ids": array(enum(ids(KindMobility, KindRecovery))),
	})
	b, err := json.Marshal(plan)
	if err != nil {
		panic(err)
	}
	return b
})

// ReasonSchemaName is the name and the version of the schema of the
// reviser. For a revision, Luna writes the reason alone, and the rules
// give each target (D-288).
const ReasonSchemaName = "luna_reason_v1"

// ReasonSchema gives the strict JSON schema of the output of the
// reviser: one reason for each exercise, with the logged sets that it
// names. A set has its kind and its number inside its kind, from 1. The
// schema holds no target, so the reviser can not change one.
func ReasonSchema() []byte { return reasonSchemaOnce() }

var reasonSchemaOnce = sync.OnceValue(func() []byte {
	var strength []string
	for _, e := range domain.DefaultCatalog().Exercises {
		if e.Kind != domain.KindCardio {
			strength = append(strength, string(e.ID))
		}
	}
	set := object(obj{
		"kind":   enum([]string{string(domain.SetWorking), string(domain.SetCalibration)}),
		"number": typ("integer"),
	})
	reason := object(obj{
		"exercise_id": enum(strength),
		"logged_sets": array(set),
		"reason":      typ("string"),
	})
	b, err := json.Marshal(object(obj{"reasons": array(reason)}))
	if err != nil {
		panic(err)
	}
	return b
})

// schemaOf gives the schema name and the schema of a role.
func schemaOf(r Role) (string, []byte) {
	if r.Name == RoleReviser {
		return ReasonSchemaName, ReasonSchema()
	}
	return SchemaName, Schema()
}

type obj = map[string]any

func typ(t string) obj { return obj{"type": t} }

func enum(values []string) obj { return obj{"type": "string", "enum": values} }

func array(items obj) obj { return obj{"type": "array", "items": items} }

// object gives a strict object: each property is required.
func object(props obj) obj {
	var req []string
	for k := range props {
		req = append(req, k)
	}
	slices.Sort(req)
	return obj{"type": "object", "additionalProperties": false, "required": req, "properties": props}
}
