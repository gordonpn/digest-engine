package httpserver

import (
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/livez", 200}, {"POST", "/livez", 405}, {"GET", "/", 404}, {"GET", "/livez/extra", 404},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			NewHandler().ServeHTTP(response, httptest.NewRequest(tc.method, tc.path, nil))
			if response.Code != tc.status {
				t.Fatalf("status = %d; want %d", response.Code, tc.status)
			}
			if tc.status == 200 && response.Body.String() != "ok\n" {
				t.Fatalf("unexpected body %q", response.Body.String())
			}
		})
	}
}
