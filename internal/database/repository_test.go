package database

import (
	"testing"

	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestGetVersion(t *testing.T) {
	tests := []struct{
		name							string
		expectedResponse  *string
		expectedError     error
	} {
		{
			name: "get_version",
			expectedResponse: new("3.53.4"),
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

			response, err := repository.GetVersion(ctx)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if *test.expectedResponse != *response {
				t.Fatalf("unexpected response, %s got %s", *test.expectedResponse, *response)
			}
		})
	}
}
 
