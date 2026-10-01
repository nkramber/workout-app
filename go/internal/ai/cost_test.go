package ai

import (
	"errors"
	"sync"
	"testing"
)

func TestNanoUSDString(t *testing.T) {
	for n, want := range map[NanoUSD]string{
		0: "0 USD", USD: "1 USD", 1_025_000: "0.001025 USD", 2*USD + 500_000_000: "2.5 USD", -1: "-0.000000001 USD",
	} {
		if got := n.String(); got != want {
			t.Errorf("%d: %q: want %q", int64(n), got, want)
		}
	}
}

func TestCapsFromEnv(t *testing.T) {
	env := func(user, project string) func(string) string {
		return func(k string) string {
			return map[string]string{EnvUserCap: user, EnvProjectCap: project}[k]
		}
	}
	c, err := CapsFromEnv(env("2", "10.5"))
	if err != nil || c != (Caps{2 * USD, 10*USD + USD/2}) {
		t.Fatalf("caps %+v, err %v", c, err)
	}
	c, err = CapsFromEnv(env("0", "0.000000001"))
	if err != nil || c != (Caps{0, 1}) {
		t.Fatalf("caps %+v, err %v", c, err)
	}
	for _, tc := range [][2]string{
		{"", "1"}, {"1", ""}, {"-1", "1"}, {"1", "abc"}, {".5", "1"}, {"1", "0.0000000001"},
		{"1e3", "1"}, {"1,5", "1"}, {"1", "2000000"}, {"1.5.5", "1"}, {" 1", "1"},
	} {
		if _, err := CapsFromEnv(env(tc[0], tc[1])); err == nil {
			t.Errorf("caps %q %q: no error", tc[0], tc[1])
		}
	}
}

func TestMemoryCap(t *testing.T) {
	m := NewMemoryCap(Caps{User: 100, Project: 150})
	settle, err := m.Reserve("a", 80)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Reserve("a", 30); !errors.Is(err, ErrCap) {
		t.Fatalf("err %v: the user cap holds the reservation", err)
	}
	if _, err := m.Reserve("b", 80); !errors.Is(err, ErrCap) {
		t.Fatalf("err %v: the project cap holds the reservation", err)
	}
	settle(20)
	settle(70) // a second settle changes nothing
	if u, p := m.Spent("a"); u != 20 || p != 20 {
		t.Fatalf("spent %d %d: want 20 20", u, p)
	}
	if _, err := m.Reserve("b", 100); err != nil {
		t.Fatalf("err %v", err)
	}
	if _, err := m.Reserve("a", 31); !errors.Is(err, ErrCap) {
		t.Fatalf("err %v: the project cap holds 120 of 150", err)
	}
}

// TestMemoryCapParallel: the spend never passes the cap with many
// goroutines.
func TestMemoryCapParallel(t *testing.T) {
	m := NewMemoryCap(Caps{User: 1000, Project: 1000})
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for range 100 {
		wg.Go(func() {
			if settle, err := m.Reserve("a", 30); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
				settle(30)
			}
		})
	}
	wg.Wait()
	if u, _ := m.Spent("a"); ok != 33 || u != 990 {
		t.Fatalf("%d calls and %d spent: want 33 and 990", ok, u)
	}
}
