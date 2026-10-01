package ai

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// NanoUSD is an amount of money in billionths of a US dollar. An
// integer count keeps each price and each cost exact.
type NanoUSD int64

// USD is one US dollar.
const USD NanoUSD = 1_000_000_000

// String gives the amount in US dollars, such as "0.001025 USD".
func (n NanoUSD) String() string {
	sign := ""
	if n < 0 {
		sign, n = "-", -n
	}
	frac := strings.TrimRight(fmt.Sprintf("%09d", int64(n%USD)), "0")
	if frac == "" {
		return fmt.Sprintf("%s%d USD", sign, int64(n/USD))
	}
	return fmt.Sprintf("%s%d.%s USD", sign, int64(n/USD), frac)
}

// Prices are the prices of one model for each token.
type Prices struct {
	Input       NanoUSD
	CachedInput NanoUSD
	Output      NanoUSD
}

// Usage is the token count of one call. The cached input tokens are
// part of the input tokens, and the reasoning tokens are part of the
// output tokens.
type Usage struct {
	InputTokens       int64
	CachedInputTokens int64
	OutputTokens      int64
	ReasoningTokens   int64
}

// Cost gives the cost of a usage. A reasoning token bills as an output
// token.
func (p Prices) Cost(u Usage) NanoUSD {
	fresh := u.InputTokens - u.CachedInputTokens
	return NanoUSD(fresh)*p.Input + NanoUSD(u.CachedInputTokens)*p.CachedInput + NanoUSD(u.OutputTokens)*p.Output
}

// worst gives the largest cost of a call with a request of n bytes. One
// token holds one byte or more, so n bytes give n input tokens or fewer.
func (r Role) worst(n int) NanoUSD {
	return NanoUSD(n)*r.Prices.Input + NanoUSD(r.MaxOutputTokens)*r.Prices.Output
}

// CostRecord is the cost of one call (D-25). It holds ids and numbers
// alone, so it can go into a log (D-80). Reserved is the worst-case
// cost that the cap hook reserved. Cost is the cost from the usage.
// When the call failed, the charge is not known, and Cost is the
// reserved cost.
type CostRecord struct {
	User       string
	Role       RoleName
	Model      string
	Effort     string
	PromptHash string
	Status     Status
	Usage      Usage
	Reserved   NanoUSD
	Cost       NanoUSD
	Known      bool
}

// ErrCap is the error of a call that the cap can not cover (D-25).
var ErrCap = errors.New("ai: the cap can not cover the call")

// CapHook reserves the worst-case cost of a call before the call
// (D-25). Reserve gives ErrCap for a call over the cap, and the client
// then makes no call. After the call, the client gives the cost to
// settle, one time.
type CapHook interface {
	Reserve(user string, worst NanoUSD) (settle func(cost NanoUSD), err error)
}

// Caps holds the cap of one user and the cap of the project (D-25).
// Q-98 gives the values in Phase 4, so they come from the configuration.
type Caps struct {
	User    NanoUSD
	Project NanoUSD
}

// The names of the configuration values of the caps, in US dollars.
const (
	EnvUserCap    = "LUNA_CAP_USER_USD"
	EnvProjectCap = "LUNA_CAP_PROJECT_USD"
)

// CapsFromEnv reads the caps from the configuration. It refuses a value
// that is not set, so a missing cap stops the start and never permits
// a call with no cap. A cap of 0 refuses each call.
func CapsFromEnv(getenv func(string) string) (Caps, error) {
	user, err := parseUSD(EnvUserCap, getenv(EnvUserCap))
	if err != nil {
		return Caps{}, err
	}
	project, err := parseUSD(EnvProjectCap, getenv(EnvProjectCap))
	if err != nil {
		return Caps{}, err
	}
	return Caps{user, project}, nil
}

// parseUSD reads a decimal amount of US dollars, such as "2" or "0.25",
// with 9 digits after the point at most.
func parseUSD(name, s string) (NanoUSD, error) {
	if s == "" {
		return 0, fmt.Errorf("ai: %s is not set", name)
	}
	whole, frac, _ := strings.Cut(s, ".")
	if whole == "" || len(frac) > 9 || !digits(whole) || !digits(frac) {
		return 0, fmt.Errorf("ai: %s: want an amount of US dollars, such as 2 or 0.25", name)
	}
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || w > int64(1_000_000) {
		return 0, fmt.Errorf("ai: %s: want 1000000 USD or less", name)
	}
	f := int64(0)
	if frac != "" {
		f, _ = strconv.ParseInt(frac+strings.Repeat("0", 9-len(frac)), 10, 64)
	}
	return NanoUSD(w)*USD + NanoUSD(f), nil
}

func digits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// MemoryCap is a cap hook that holds the spend in memory, for one
// period. Phase 4 gives the store and the period of the caps (Q-98).
// It is safe for use by more than one goroutine.
type MemoryCap struct {
	mu      sync.Mutex
	caps    Caps
	user    map[string]NanoUSD
	project NanoUSD
}

// NewMemoryCap gives a cap hook with no spend.
func NewMemoryCap(c Caps) *MemoryCap {
	return &MemoryCap{caps: c, user: map[string]NanoUSD{}}
}

// Reserve reserves the worst-case cost of a call, or gives ErrCap when
// the spend and the reservations of the user or of the project can not
// cover it.
func (m *MemoryCap) Reserve(user string, worst NanoUSD) (func(NanoUSD), error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.user[user]+worst > m.caps.User || m.project+worst > m.caps.Project {
		return nil, ErrCap
	}
	m.user[user] += worst
	m.project += worst
	var once sync.Once
	return func(cost NanoUSD) {
		once.Do(func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.user[user] += cost - worst
			m.project += cost - worst
		})
	}, nil
}

// Spent gives the spend and the open reservations of a user and of the
// project.
func (m *MemoryCap) Spent(user string) (NanoUSD, NanoUSD) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.user[user], m.project
}
