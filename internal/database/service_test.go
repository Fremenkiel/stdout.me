package database

import (
	"context"
	"testing"

	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestGetStats(t *testing.T) {
	tests := []struct{
		name							string
		expectedResponse  Stats
		expectedError     error
	} {
		{
			name: "get_stats",
			expectedResponse: Stats{
				Version: "0.1.0",
				Size: 24.576,
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

			service := NewService(
				&mockRepository{},
				sessionService,
				)

			response, err := service.GetStats(ctx)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if test.expectedError != nil {
				return
			}

			if test.expectedResponse.Size != response.Size {
				t.Fatalf("unexpected response size, %v got %v", test.expectedResponse.Size, response.Size)
			}

			if test.expectedResponse.Version != response.Version {
				t.Fatalf("unexpected response size, %s got %s", test.expectedResponse.Version, response.Version)
			}
		})
	}
}

type mockRepository struct {}

func (r *mockRepository) GetVersion(ctx context.Context) (*string, error) {
	return new("0.1.0"), nil
}
