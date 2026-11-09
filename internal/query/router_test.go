package query

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestNewRouter(t *testing.T) {
	tests := []struct{
		name									string
		url										string
		expectedResponse			string
		expectedResponseCode  int
	} {
		{
			name: "index",
			url: "/",
			expectedResponse: "page:index",
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "query",
			url: "/query?sql=select%20name%20from%20users%3B%0A",
			expectedResponse: "fragment:result-swap",
			expectedResponseCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()

			handler := &Handler{
				testutil.NewRenderer(),
				&mockService{},
			}

			NewRouter(mux, handler)

			req := httptest.NewRequestWithContext(
				context.WithValue(t.Context(), "session_key", "test"),
				http.MethodGet, test.url, nil)

			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if test.expectedResponseCode != rr.Code {
				t.Fatalf("unexpected status code, %d got %d", http.StatusOK, rr.Code)
			}

			if !strings.Contains(rr.Body.String(), test.expectedResponse) {
				t.Fatalf("unexpected response body, %s", rr.Body.String())
			}
		})
	}
}

