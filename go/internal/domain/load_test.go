package domain

import (
	"errors"
	"testing"
)

func TestLoadString(t *testing.T) {
	for _, tc := range []struct {
		load Load
		want string
	}{
		{Pounds(40), "40 lb"},
		{125, "12.5 lb"},
		{25, "2.5 lb"},
		{Tenth, "0.1 lb"},
		{0, "0 lb"},
		{-5, "-0.5 lb"},
		{-Pounds(5), "-5 lb"},
	} {
		if got := tc.load.String(); got != tc.want {
			t.Errorf("Load(%d).String() = %q, want %q", tc.load, got, tc.want)
		}
	}
}

func TestLoadExact(t *testing.T) {
	// 12.5 lb steps stay exact over a long stack (D-160).
	var sum Load
	for range 40 {
		sum += 125
	}
	if sum != Pounds(500) {
		t.Fatalf("40 x 12.5 lb = %s, want 500 lb", sum)
	}
}

func TestLoadCheck(t *testing.T) {
	for _, tc := range []struct {
		load Load
		ok   bool
	}{
		{Tenth, true},
		{125, true},
		{Pounds(400), true},
		{0, false},
		{-Pound, false},
	} {
		err := tc.load.Check()
		if (err == nil) != tc.ok {
			t.Errorf("Load(%d).Check() = %v, want ok %v", tc.load, err, tc.ok)
		}
		if err != nil && !errors.Is(err, ErrInvalid) {
			t.Errorf("Load(%d).Check() = %v, want ErrInvalid", tc.load, err)
		}
	}
}
