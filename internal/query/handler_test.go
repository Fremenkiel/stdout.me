package query

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fremenkiel/stdout.me/internal/database"
	"github.com/fremenkiel/stdout.me/internal/registry"
	"github.com/fremenkiel/stdout.me/internal/schema"
	"github.com/fremenkiel/stdout.me/internal/session"
	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestHandlerIndex(t *testing.T) {
	tests := []struct{
		name,
		method,
		url											string
		hasSession							bool
		expectedResponseContent	[]string
		expectedResponseCode		int
	} {
		{
			name: "index",
			method: http.MethodGet,
			url: "/",
			hasSession: true,
			expectedResponseContent: []string {
				"page:index",
				"test_table",
				"876",
			},
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "no_session",
			method: http.MethodGet,
			url: "/",
			hasSession: false,
			expectedResponseContent: nil,
			expectedResponseCode: http.StatusInternalServerError,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.hasSession {
				ctx = context.WithValue(ctx, "session_key", "test")
			}

			handler := NewHandler(
				testutil.NewRenderer(),
				&mockService{},
				)

			rr := httptest.NewRecorder()

			handler.Index(
				rr,
				httptest.NewRequestWithContext(ctx, test.method, test.url, nil),
				)

			if test.expectedResponseCode != rr.Code {
				t.Fatalf("unexpected status code, %d got %d", test.expectedResponseCode, rr.Code)
			}

			for _, content := range test.expectedResponseContent {
				if !strings.Contains(rr.Body.String(), content) {
					t.Fatalf("unexpected response body, %s got %s", content, rr.Body.String())
				}
			}
		})
	}
}

func TestHandlerRun(t *testing.T) {
	tests := []struct{
		name,
		method,
		url											string
		hasSession							bool
		expectedResponseContent	[]string
		expectedResponse				int
	} {
		{
			name: "get",
			method: http.MethodGet,
			url: "/query?sql=select%20name%20from%20users%3B%0A",
			hasSession: true,
			expectedResponseContent: []string {
				"fragment:result",
				"test_row",
				"test_col",
			},
			expectedResponse: http.StatusOK,
		},
		{
			name: "get_no_query",
			method: http.MethodGet,
			url: "/query",
			hasSession: true,
			expectedResponseContent: nil,
			expectedResponse: http.StatusNoContent,
		},
		{
			name: "get_no_return",
			method: http.MethodGet,
			url: "/query?sql=delete%20from%20users%3B%0A",
			hasSession: true,
			expectedResponseContent: nil,
			expectedResponse: http.StatusOK,
		},
		{
			name: "get_no_session",
			method: http.MethodGet,
			url: "/query?sql=select%20name%20from%20users%3B%0A",
			hasSession: false,
			expectedResponseContent: nil,
			expectedResponse: http.StatusInternalServerError,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.hasSession {
				ctx = context.WithValue(ctx, "session_key", "test")
			}

			handler := NewHandler(
				testutil.NewRenderer(),
				&mockService{
					sleep: time.Duration(0),
				},
				)

			rr := httptest.NewRecorder()

			handler.Run(
				rr,
				httptest.NewRequestWithContext(ctx, test.method, test.url, nil),
				)

			if test.expectedResponse != rr.Code {
				t.Fatalf("unexpected status code, %d got %d", test.expectedResponse, rr.Code)
			}

			for _, content := range test.expectedResponseContent {
				if !strings.Contains(rr.Body.String(), content) {
					t.Fatalf("unexpected response body, %s got %s", content, rr.Body.String())
				}
			}
		})
	}
}

func TestHandlerRunDuration(t *testing.T) {
	tests := []struct{
		name,
		method,
		url								string
		hasSession				bool
		expectedDuration	time.Duration
	} {
		{
			name: "get",
			method: http.MethodGet,
			url: "/query?sql=select%20name%20from%20users%3B%0A",
			hasSession: true,
			expectedDuration: time.Duration(10 * time.Microsecond),
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := t.Context()
				if test.hasSession {
					ctx = context.WithValue(ctx, "session_key", "test")
				}

				handler := NewHandler(
					testutil.NewRenderer(),
					&mockService{
						sleep: test.expectedDuration,
					},
					)

				rr := httptest.NewRecorder()

				handler.Run(
					rr,
					httptest.NewRequestWithContext(ctx, test.method, test.url, nil),
					)

				expectedContent := fmt.Sprintf("%v μs", test.expectedDuration.Microseconds())
				if !strings.Contains(rr.Body.String(), expectedContent) {
					t.Fatalf("unexpected response body, %s got %s", expectedContent, rr.Body.String())
				}
			})
		})
	}
}

type mockService struct {
	sleep time.Duration
}

var _ service = (*mockService)(nil)

func (m *mockService) GetInfo(ctx context.Context) (*Query, error) {
	_, ok := ctx.Value("session_key").(string)
	if !ok {
		return nil, session.ErrNoSessionInContext
	}

	time.Sleep(m.sleep)

	schema, err := m.getSchema(ctx)
	if err != nil {
		return nil, err
	}

	state, err := m.getStats(ctx)
	if err != nil {
		return nil, err
	}

	return &Query{
		Schema: schema,
		Stats: state,
		Result: nil,
		Registry: &registry.Registry{
			Lookup: map[string]int{
				"test_table": 1,
			},
		},
	}, nil
}

func (m *mockService) Run(ctx context.Context, query string) (*Query, error) {
	_, ok := ctx.Value("session_key").(string)
	if !ok {
		return nil, session.ErrNoSessionInContext
	}

	time.Sleep(m.sleep)

	return &Query{
		Schema: nil,
		Stats: nil,
		Result: &Result{
			Rows: []*Row{
				{
					new("test_row"),
				},
			},
			Columns: []string{
				"test_col",
			},
			Duration: 765,
		},
	}, nil
}

func (m *mockService) getSchema(ctx context.Context) (schema.Schema, error) {
	_, ok := ctx.Value("session_key").(string)
	if !ok {
		return nil, session.ErrNoSessionInContext
	}

	return schema.Schema{
		{
			Name: "test_table",
			RowCount: 876,
			Columns: []*schema.Column{
				{
					Name: "test_col",
					Type: "INTEGER",
				},
			},
		},
	}, nil
}

func (m *mockService) getStats(ctx context.Context) (*database.Stats, error) {
	_, ok := ctx.Value("session_key").(string)
	if !ok {
		return nil, session.ErrNoSessionInContext
	}

	return &database.Stats{
		Version: "0.1.0",
		Size: 100,
	}, nil
}
