package auth

import "testing"

func TestBearer(t *testing.T) {
	cases := map[string]string{
		"":                   "",
		"Bearer":             "",
		"Bearer ":            "",
		"Bearer abc":         "abc",
		"bearer abc":         "abc",
		"BEARER  abc ":       "abc",
		"Bearer\tabc":        "",
		"Bearer a b":         "",
		"Basic abc":          "",
		"Bearerabc":          "",
		"  Bearer abc.def.g": "abc.def.g",
	}
	for in, want := range cases {
		if got := bearer(in); got != want {
			t.Errorf("bearer(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseOrigin(t *testing.T) {
	good := []string{"", "https://nk-workout-app-prod.web.app", "http://127.0.0.1:5173"}
	for _, v := range good {
		if got, err := ParseOrigin(v); err != nil || got != v {
			t.Errorf("ParseOrigin(%q) = %q, %v, want the value", v, got, err)
		}
	}
	bad := []string{"*", "nk-workout-app-prod.web.app", "https://a.web.app/", "https://a.web.app,https://b.web.app",
		"https://a.web.app?x=1", "https://user@a.web.app", "ftp://a.web.app"}
	for _, v := range bad {
		if _, err := ParseOrigin(v); err == nil {
			t.Errorf("ParseOrigin(%q) = nil error, want a refusal", v)
		}
	}
}
