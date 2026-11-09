package schema

import (
	"testing"

	"github.com/fremenkiel/stdout.me/internal/session"
	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestRepositoryGetSchema(t *testing.T) {
	tests := []struct{
		name							string
		hasSession				bool
		expectedResponse	Schema
		expectedError			error
	} {
		{
			name: "get_tables",
			hasSession: true,
			expectedResponse: Schema{
				{
					Name: "educations",
					RowCount: 0,
					Columns: []*Column{
						{
							Name: "id",
							Type: "INTEGER",
							IsNullable: false,
							DefaultValue: "",
							IsPrimaryKey: true,
							IsForeignKey: false,
						},
						{
							Name: "user_id",
							Type: "INTEGER",
							IsNullable: false,
							DefaultValue: "",
							IsPrimaryKey: false,
							IsForeignKey: true,
						},
						{
							Name: "description",
							Type: "TEXT",
							IsNullable: false,
							DefaultValue: "",
							IsPrimaryKey: false,
							IsForeignKey: false,
						},
					},
				},
				{
					Name: "users",
					RowCount: 1,
					Columns: []*Column{
						{
							Name: "id",
							Type: "INTEGER",
							IsNullable: false,
							DefaultValue: "",
							IsPrimaryKey: true,
							IsForeignKey: false,
						},
						{
							Name: "name",
							Type: "TEXT",
							IsNullable: true,
							DefaultValue: "",
							IsPrimaryKey: false,
							IsForeignKey: false,
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "no_session",
			hasSession: false,
			expectedResponse: nil,
			expectedError: session.ErrNoSessionInContext,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			sessionService := testutil.NewMockSessionService(t)

			ctx := t.Context()
			defer sessionService.CleanupSession(ctx)
			if test.hasSession {
				sessionCtx, err := sessionService.CreateSessionContext(ctx)
				if err != nil {
					t.Fatal(err)
				}

				ctx = sessionCtx
			}

			repository := NewRepository(sessionService.Cache)

			response, err := repository.GetSchema(ctx)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if test.expectedError != nil && err != nil {
				return
			}

			if len(test.expectedResponse) != len(response) {
				t.Fatalf("unexpected response len, %d got %d", len(test.expectedResponse), len(response))
			}

			for i := range response {
				table := response[i]
				expectedResponse := test.expectedResponse[i]

				if expectedResponse.Name != table.Name {
					t.Fatalf("unexpected name, %s got %s", expectedResponse.Name, table.Name)
				}

				if expectedResponse.RowCount != table.RowCount {
					t.Fatalf("unexpected name, %s got %s", expectedResponse.Name, table.Name)
				}

				for ci := range table.Columns {
					col := table.Columns[i]
					expectedColumn := expectedResponse.Columns[ci]

					if expectedColumn.Name != col.Name {
						t.Fatalf("unexpected column name, %s got %s", expectedColumn.Name, col.Name)
					}

					if expectedColumn.Type != col.Type {
						t.Fatalf("unexpected type for %s, %s got %s", col.Name, expectedColumn.Type, col.Type)
					}

					if expectedColumn.IsNullable != col.IsNullable {
						t.Fatalf("unexpected nullable state for %s, %v got %v", col.Name, expectedColumn.IsNullable, col.IsNullable)
					}

					if expectedColumn.DefaultValue != col.DefaultValue {
						t.Fatalf("unexpected default value for %s, '%s' got '%s'", col.Name, expectedColumn.DefaultValue, col.DefaultValue)
					}

					if expectedColumn.IsPrimaryKey != col.IsPrimaryKey {
						t.Fatalf("unexpected primary key state for %s, %v got %v", col.Name, expectedColumn.IsPrimaryKey, col.IsPrimaryKey)
					}

					if expectedColumn.IsForeignKey != col.IsForeignKey {
						t.Fatalf("unexpected foreign key state for %s, %v got %v", col.Name, expectedColumn.IsForeignKey, col.IsForeignKey)
					}
				}
			}

		})
	}
}

func TestRepositoryGetColumns(t *testing.T) {
	tests := []struct{
		name,
		tableName					string
		hasSession				bool
		expectedResponse	[]*Column
		expectedError			error
	} {
		{
			name: "get_tables",
			tableName: "educations",
			hasSession: true,
			expectedResponse: []*Column{
				{
					Name: "id",
					Type: "INTEGER",
					IsNullable: false,
					DefaultValue: "",
					IsPrimaryKey: true,
					IsForeignKey: false,
				},
				{
					Name: "user_id",
					Type: "INTEGER",
					IsNullable: false,
					DefaultValue: "",
					IsPrimaryKey: false,
					IsForeignKey: true,
				},
				{
					Name: "description",
					Type: "TEXT",
					IsNullable: false,
					DefaultValue: "",
					IsPrimaryKey: false,
					IsForeignKey: false,
				},
			},
			expectedError: nil,
		},
		{
			name: "no_session",
			hasSession: false,
			expectedResponse: nil,
			expectedError: session.ErrNoSessionInContext,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			sessionService := testutil.NewMockSessionService(t)

			ctx := t.Context()
			defer sessionService.CleanupSession(ctx)
			if test.hasSession {
				sessionCtx, err := sessionService.CreateSessionContext(ctx)
				if err != nil {
					t.Fatal(err)
				}

				ctx = sessionCtx
			}

			repository := NewRepository(sessionService.Cache)

			response, err := repository.GetColumns(ctx, test.tableName)
			if test.expectedError != err {
				t.Fatalf("unexpected err, %v got %v", test.expectedError, err)
			}

			if test.expectedError != nil && err != nil {
				return
			}

			if len(test.expectedResponse) != len(response) {
				t.Fatalf("unexpected response len, %d got %d", len(test.expectedResponse), len(response))
			}


			for i := range response {
				col := response[i]
				expectedResponse := test.expectedResponse[i]

				if expectedResponse.Name != col.Name {
					t.Fatalf("unexpected column name, %s got %s", expectedResponse.Name, col.Name)
				}

				if expectedResponse.Type != col.Type {
					t.Fatalf("unexpected type for %s, %s got %s", col.Name, expectedResponse.Type, col.Type)
				}

				if expectedResponse.IsNullable != col.IsNullable {
					t.Fatalf("unexpected nullable state for %s, %v got %v", col.Name, expectedResponse.IsNullable, col.IsNullable)
				}

				if expectedResponse.DefaultValue != col.DefaultValue {
					t.Fatalf("unexpected default value for %s, '%s' got '%s'", col.Name, expectedResponse.DefaultValue, col.DefaultValue)
				}

				if expectedResponse.IsPrimaryKey != col.IsPrimaryKey {
					t.Fatalf("unexpected primary key state for %s, %v got %v", col.Name, expectedResponse.IsPrimaryKey, col.IsPrimaryKey)
				}

				if expectedResponse.IsForeignKey != col.IsForeignKey {
					t.Fatalf("unexpected foreign key state for %s, %v got %v", col.Name, expectedResponse.IsForeignKey, col.IsForeignKey)
				}
			}
		})
	}
}
