package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testKey = "sk-test-key-not-real"

// server gives a local server in place of OpenAI. No test calls OpenAI.
func server(t *testing.T, status int, body string, check func(r *http.Request, b map[string]any)) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var b map[string]any
		_ = json.Unmarshal(raw, &b)
		if check != nil {
			check(r, b)
		}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(s.Close)
	return s
}

func testCall(t *testing.T) Call {
	input, err := userInput(request(t))
	if err != nil {
		t.Fatal(err)
	}
	return Call{Role: Planner(), Instructions: Instructions(Planner()), Input: input, SchemaName: SchemaName, Schema: Schema()}
}

func TestOpenAIRequest(t *testing.T) {
	body := `{"status":"completed","output":[{"type":"reasoning"},{"type":"message","content":[{"type":"output_text","text":"{\"a\":"},{"type":"output_text","text":"1}"}]}],
		"usage":{"input_tokens":900,"input_tokens_details":{"cached_tokens":600,"cache_write_tokens":200},"output_tokens":300,"output_tokens_details":{"reasoning_tokens":100}}}`
	c := testCall(t)
	s := server(t, 200, body, func(r *http.Request, b map[string]any) {
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer "+testKey || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("request %s %v", r.Method, r.Header)
		}
		if b["model"] != Planner().Model || b["reasoning"].(map[string]any)["effort"] != "medium" || b["store"] != false {
			t.Errorf("body model %v reasoning %v store %v", b["model"], b["reasoning"], b["store"])
		}
		if b["instructions"] != c.Instructions || b["input"] != string(c.Input) || b["max_output_tokens"] != float64(Planner().MaxOutputTokens) {
			t.Error("the body does not hold the call")
		}
		f := b["text"].(map[string]any)["format"].(map[string]any)
		if f["type"] != "json_schema" || f["name"] != SchemaName || f["strict"] != true || f["schema"] == nil {
			t.Errorf("format %v", f)
		}
	})
	o, err := NewOpenAI(testKey, s.URL, s.Client())
	if err != nil {
		t.Fatal(err)
	}
	r, err := o.Send(context.Background(), c)
	if err != nil {
		t.Fatal(err)
	}
	if r.Text != `{"a":1}` || r.Refusal || r.Incomplete || r.Usage != (Usage{900, 600, 200, 300, 100}) {
		t.Fatalf("reply %+v", r)
	}
}

func TestOpenAIReplies(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body       string
		refusal    bool
		incomplete bool
		err        string
	}{
		{"refusal", 200, `{"status":"completed","output":[{"type":"message","content":[{"type":"refusal","refusal":"no"}]}]}`, true, false, ""},
		{"incomplete", 200, `{"status":"incomplete","output":[]}`, false, true, ""},
		{"failed", 200, `{"status":"failed","output":[]}`, false, false, "did not complete"},
		{"not JSON", 200, `<html>`, false, false, "not JSON"},
		{"HTTP 429", 429, `{"error":{"message":"` + testKey + `"}}`, false, false, "HTTP 429"},
		{"HTTP 500", 500, ``, false, false, "HTTP 500"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := server(t, tc.status, tc.body, nil)
			o, _ := NewOpenAI(testKey, s.URL, s.Client())
			r, err := o.Send(context.Background(), testCall(t))
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("err %v: want %q", err, tc.err)
				}
				if strings.Contains(err.Error(), testKey) {
					t.Fatal("the error holds the key")
				}
				return
			}
			if err != nil || r.Refusal != tc.refusal || r.Incomplete != tc.incomplete {
				t.Fatalf("reply %+v err %v", r, err)
			}
		})
	}
}

// TestOpenAITimeout: a slow server gives the error of the context, and
// the client reads a time-out.
func TestOpenAITimeout(t *testing.T) {
	done := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-done:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(done); s.Close() })
	o, _ := NewOpenAI(testKey, s.URL, s.Client())
	res, err := (&Client{Provider: o, Cap: bigCap(), Timeout: 50 * time.Millisecond}).Plan(context.Background(), request(t))
	if err != nil || res.Status != StatusTimeout {
		t.Fatalf("status %q err %v: want %q", res.Status, err, StatusTimeout)
	}
}

func TestNewOpenAI(t *testing.T) {
	if _, err := NewOpenAI("", "", nil); err == nil {
		t.Fatal("an empty key gives no error")
	}
	o, err := NewOpenAI(testKey, "", nil)
	if err != nil || o.endpoint != OpenAIEndpoint || o.client != http.DefaultClient {
		t.Fatalf("provider %+v err %v", o, err)
	}
	o, _ = NewOpenAI(testKey, "http://bad host", nil)
	if _, err := o.Send(context.Background(), testCall(t)); err == nil || strings.Contains(err.Error(), testKey) {
		t.Fatalf("err %v", err)
	}
}

// TestOpenAICacheWriteCost: a reply with cache-write tokens costs the
// published rates of the planner model, read 2026-10-01: per million tokens,
// 0.10 USD input, 0.01 USD cached input, 0.125 USD cache write, and
// 0.50 USD output.
func TestOpenAICacheWriteCost(t *testing.T) {
	body := `{"status":"completed","output":[],"usage":{"input_tokens":1000000,"input_tokens_details":{"cached_tokens":250000,"cache_write_tokens":500000},"output_tokens":1000000}}`
	s := server(t, 200, body, nil)
	o, _ := NewOpenAI(testKey, s.URL, s.Client())
	r, err := o.Send(context.Background(), testCall(t))
	if err != nil {
		t.Fatal(err)
	}
	// 250,000 ordinary input tokens, 250,000 cache reads, 500,000 cache
	// writes, and 1,000,000 output tokens.
	want := USD/40 + USD/400 + USD/16 + USD/2
	if got := Planner().Prices.Cost(r.Usage); got != want {
		t.Fatalf("cost %s: want %s", got, want)
	}
}
