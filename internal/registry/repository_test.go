package registry

import (
	"errors"
	"testing"

	"github.com/fremenkiel/stdout.me/internal/testutil"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func TestRegistryGetAll(t *testing.T) {
	tests := []struct{
		name							string
		expectedResponse  *Registry
		expectedError     error
	} {
		{
			name: "get_all",
			expectedResponse: &Registry{
				Lookup: map[string]int{
					"users": 1,
					"educations": 2,
				},
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sessionService := testutil.NewMockSessionService(t)

			ctx, err := sessionService.CreateSessionContext(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer sessionService.CleanupSession(ctx)

			repository := NewRepository(sessionService.Cache)

			response, err := repository.GetAll(ctx)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if test.expectedResponse == nil && response != nil {
				t.Fatalf("unexpected response, %v got %v", test.expectedResponse, response)
			}

			if len(test.expectedResponse.Lookup) != len(response.Lookup) {
				t.Fatalf("unexpected response len, %v got %v", len(test.expectedResponse.Lookup), len(response.Lookup))
			}

			for key, val := range response.Lookup {
				row, ok := test.expectedResponse.Lookup[key]
				if !ok {
					t.Fatalf("unexpected key, %s does not exist", key)
				}

				if row != val {
					t.Fatalf("unexpected value, %v got %v (%s)", row, val, key)
				}
			}
		})
	}
}

func TestRegistryCreate(t *testing.T) {
	tests := []struct{
		name							string
		tableName					string
		expectedResponse  *int
		expectedError     error
	} {
		{
			name: "create",
			tableName: "accounts",
			expectedResponse: new(3),
			expectedError: nil,
		},
		{
			name: "quoted_create",
			tableName: "[account list]",
			expectedResponse: new(3),
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sessionService := testutil.NewMockSessionService(t)

			ctx, err := sessionService.CreateSessionContext(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer sessionService.CleanupSession(ctx)

			repository := NewRepository(sessionService.Cache)

			response, err := repository.Create(ctx, test.tableName)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if *test.expectedResponse != *response {
				t.Fatalf("unexpected response, %d got %d", *test.expectedResponse, *response)
			}
		})
	}
}

func TestRegistryUpdate(t *testing.T) {
	tests := []struct{
		name							string
		id								int
		tableName					string
		expectedError     error
	} {
		{
			name: "update",
			id: 1,
			tableName: "accounts",
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sessionService := testutil.NewMockSessionService(t)

			ctx, err := sessionService.CreateSessionContext(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer sessionService.CleanupSession(ctx)

			repository := NewRepository(sessionService.Cache)

			if err := repository.Update(ctx, test.id, test.tableName); test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}
		})
	}
}

func TestRegistryRemove(t *testing.T) {
	tests := []struct{
		name							string
		id								int
		expectedError     error
	} {
		{
			name: "remove",
			id: 1,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sessionService := testutil.NewMockSessionService(t)

			ctx, err := sessionService.CreateSessionContext(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer sessionService.CleanupSession(ctx)

			repository := NewRepository(sessionService.Cache)

			if err := repository.Remove(ctx, test.id); test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}
		})
	}
}

func TestKeyConstraint(t *testing.T) {
	sessionService := testutil.NewMockSessionService(t)

	ctx, err := sessionService.CreateSessionContext(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer sessionService.CleanupSession(ctx)

	repository := NewRepository(sessionService.Cache)

	_, err = repository.Create(ctx, "users")
	if err == nil {
		t.Fatal("unexpected err, got nil")
	}

	typedErr, ok := errors.AsType[*sqlite.Error](err)
	if !ok {
		t.Fatalf("unexpected error type, got %v", err)
	}

	if typedErr.Code() != sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		t.Fatalf("unexpected error code, got %v", typedErr.Code())
	}
}
