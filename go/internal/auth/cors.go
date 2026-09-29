package auth

import (
	"fmt"
	"net/http"
	"net/url"
)

// ParseOrigin checks the ALLOWED_ORIGIN value: one origin with a scheme
// and a host, and no path, or empty (D-82). Empty allows no cross-origin
// call.
func ParseOrigin(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" ||
		u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", fmt.Errorf("auth: ALLOWED_ORIGIN must be one origin such as https://example.web.app")
	}
	return v, nil
}

// CORS answers the cross-origin checks of the browser for the one origin
// of the web app (D-82), as decktome:go/internal/auth/cors.go does. A
// request from another origin gets no CORS header, and the browser
// refuses the answer. A preflight from another origin gets no answer of
// its own, and the handler behind it refuses the OPTIONS method.
func CORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Vary", "Origin")
		if origin != "" && r.Header.Get("Origin") == origin {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")
			h.Set("Access-Control-Expose-Headers", "Connect-Protocol-Version")
			h.Set("Access-Control-Max-Age", "600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
