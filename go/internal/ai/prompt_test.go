package ai

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/nkramber/workout-app/go/internal/policy"
)

// TestInstructions: the instructions state the boundary of D-36 with
// the dated copy of the usage policies of D-93, the limits of the text
// of Luna, each guidance item, and each rule of the policy.
func TestInstructions(t *testing.T) {
	for _, r := range Roles() {
		text := Instructions(r)
		for _, want := range []string{
			string(r.Name), PromptVersion,
			"Do not diagnose, treat, or prescribe rehabilitation",
			"no medical, emergency, diet, or weight-loss advice",
			"printed 2025-11-07 and effective 2025-10-29",
			"280 characters", "160 characters", "Write no other text",
		} {
			if !strings.Contains(text, want) {
				t.Errorf("%s: no %q", r.Name, want)
			}
		}
		for _, g := range Guidance() {
			if !strings.Contains(text, string(g.ID)) {
				t.Errorf("%s: no guidance item %q", r.Name, g.ID)
			}
		}
		for _, p := range policy.Rules() {
			if !strings.Contains(text, string(p.ID)+": ") {
				t.Errorf("%s: no policy rule %q", r.Name, p.ID)
			}
		}
		if Instructions(r) != text {
			t.Errorf("%s: the instructions change", r.Name)
		}
	}
}

func TestPromptHash(t *testing.T) {
	p, r := PromptHash(Planner()), PromptHash(Reviser())
	hex := regexp.MustCompile(`^[0-9a-f]{64}$`)
	if !hex.MatchString(p) || !hex.MatchString(r) || p == r {
		t.Fatalf("hashes %q and %q: want two different SHA-256 values", p, r)
	}
	if PromptHash(Planner()) != p {
		t.Fatal("the hash changes")
	}
}

func TestRoles(t *testing.T) {
	for _, r := range Roles() {
		if r.Model == "" || r.Effort != "medium" || r.MaxOutputTokens <= 0 || r.Timeout <= 0 || r.Prices.Output <= 0 {
			t.Errorf("role %+v", r)
		}
	}
	if Planner().MaxSessions != 7 || Reviser().MaxSessions != 1 {
		t.Fatal("the session limits changed")
	}
}

// TestNoModelAtCallSite: no Go file of the module names a model, except
// role.go of this package (D-24).
func TestNoModelAtCallSite(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	own, _ := filepath.Abs("role.go")
	model := regexp.MustCompile(`gpt-[0-9]`)
	files := 0
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		files++
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if path != own && model.Match(b) {
			t.Errorf("%s names a model: name a role of go/internal/ai", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(own)
	if files < 10 || !model.Match(b) {
		t.Fatalf("%d files, and role.go names no model: the walk is wrong", files)
	}
}
