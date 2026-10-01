package ai

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/domain"
)

// TestSchemaStrict: each object of the schema is strict. Each property
// is required, and no other property is allowed, as the structured
// outputs of OpenAI require.
func TestSchemaStrict(t *testing.T) {
	var s map[string]any
	if err := json.Unmarshal(Schema(), &s); err != nil {
		t.Fatal(err)
	}
	var walk func(path string, n map[string]any)
	objects := 0
	walk = func(path string, n map[string]any) {
		switch n["type"] {
		case "object":
			objects++
			if n["additionalProperties"] != false {
				t.Errorf("%s: additionalProperties is not false", path)
			}
			props := n["properties"].(map[string]any)
			var keys, req []string
			for k := range props {
				keys = append(keys, k)
			}
			for _, r := range n["required"].([]any) {
				req = append(req, r.(string))
			}
			slices.Sort(keys)
			if !slices.Equal(keys, req) {
				t.Errorf("%s: required %v: want %v", path, req, keys)
			}
			for k, v := range props {
				walk(path+"."+k, v.(map[string]any))
			}
		case "array":
			walk(path+"[]", n["items"].(map[string]any))
		case "string", "integer", "number":
		default:
			t.Errorf("%s: type %v", path, n["type"])
		}
	}
	walk("", s)
	if objects != 6 {
		t.Fatalf("%d objects: want 6", objects)
	}
}

// TestSchemaEnums: the enums hold each exercise of the catalog and
// each item of the guidance catalog, by kind.
func TestSchemaEnums(t *testing.T) {
	text := string(Schema())
	for _, e := range domain.DefaultCatalog().Exercises {
		if !strings.Contains(text, `"`+string(e.ID)+`"`) {
			t.Errorf("exercise %q is not in the schema", e.ID)
		}
	}
	for _, g := range Guidance() {
		if !strings.Contains(text, `"`+string(g.ID)+`"`) {
			t.Errorf("item %q is not in the schema", g.ID)
		}
	}
	if !bytes.Equal(Schema(), Schema()) {
		t.Fatal("the schema changes")
	}
}

func TestGuidanceCatalog(t *testing.T) {
	seen := map[GuidanceID]bool{}
	kinds := map[GuidanceKind]int{}
	for _, g := range Guidance() {
		if seen[g.ID] {
			t.Errorf("id %q two times", g.ID)
		}
		seen[g.ID] = true
		kinds[g.Kind]++
		if !strings.HasPrefix(string(g.ID), string(g.Kind)+".") {
			t.Errorf("id %q: want the prefix %q", g.ID, g.Kind)
		}
		if g.Text == "" || len(g.Text) > SummaryMax {
			t.Errorf("id %q: text of %d bytes", g.ID, len(g.Text))
		}
	}
	for _, k := range []GuidanceKind{KindWarmUp, KindCoolDown, KindMobility, KindRecovery} {
		if kinds[k] == 0 {
			t.Errorf("no item of kind %q", k)
		}
	}
	for id, k := range map[GuidanceID]GuidanceKind{DefaultWarmUp: KindWarmUp, DefaultCoolDown: KindCoolDown} {
		if g, ok := GuidanceItemOf(id); !ok || g.Kind != k {
			t.Errorf("default %q: item %+v", id, g)
		}
	}
	if _, ok := GuidanceItemOf("mobility.none"); ok {
		t.Error("an unknown id gives an item")
	}
	g := Guidance()
	g[0].Text = "changed"
	if Guidance()[0].Text == "changed" {
		t.Error("Guidance shares its data")
	}
}
