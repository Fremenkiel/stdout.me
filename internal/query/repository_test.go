package query

import (
	"testing"

	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestGetAll(t *testing.T) {
	tests := []struct{
		name							string
		query							string
		columns						[]string
		expectedResponse  []*Row
		expectedError     error
	} {
		{
			name: "get_all",
			query: "SELECT id, name FROM users;",
			columns: []string{
				"id",
				"name",
			},
			expectedResponse: []*Row{
				{ new("1"), new("Kevin Wagner") },
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

			response, err := repository.GetAll(ctx, test.query, test.columns)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if len(test.expectedResponse) != len(response) {
				t.Fatalf("unexpected response, %d got %d", len(test.expectedResponse), len(response))
			}

			for i := range response {
				row := response[i]
				checkRow := test.expectedResponse[i]

				if len(*checkRow) != len(*row) {
					t.Fatalf("unexpected row len, %d got %d (%d)", len(*checkRow), len(*row), i)
				}

				for ci, col := range *row {
					checkColumn := (*checkRow)[ci]
					if *checkColumn != *col {
						t.Fatalf("unexpected row column, %s got %s (%d)", *checkColumn, *col, i)
					}
				}
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct{
		name							string
		query							string
		expectedError     error
	} {
		{
			name: "execute",
			query: "UPDATE users SET name = 'James' WHERE id = 1;",
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

			if err := repository.Execute(ctx, test.query); test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}
		})
	}
}
