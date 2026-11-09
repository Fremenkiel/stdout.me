package query

import (
	"context"
	"fmt"
	"testing"

	"github.com/fremenkiel/stdout.me/internal/database"
	"github.com/fremenkiel/stdout.me/internal/registry"
	"github.com/fremenkiel/stdout.me/internal/schema"
	"github.com/fremenkiel/stdout.me/internal/session"
	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestServiceRun(t *testing.T) {
	tests := []struct{
		name,
		query							string
		expectedResponse	*Result
		expectedError			error
	} {
		{
			name: "get_rows",
			query: "users",
			expectedResponse: &Result{
				Rows: []*Row{
					{
						new("1"),
						new("Kevin"),
					},
				},	
				Columns: []string{
					"id",
					"name",
					"age",
				},
			},
			expectedError: nil,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()

			service := NewService(
				&mockRepository{},
				NewMockSchemaService(),
				NewMockDatabaseService(),
				NewMockRegistryService(),
				)

			response, err := service.Run(ctx, fmt.Sprintf("SELECT * FROM %s;", test.query))
			if test.expectedError != err {
				t.Fatalf("unexpected error, %v got %v", test.expectedError, err)
			}

			if test.expectedError != nil && err != nil {
				return
			}

			if len(test.expectedResponse.Rows) != len(response.Result.Rows) {
				t.Fatalf("unexpected row len, %d got %d", len(test.expectedResponse.Rows), len(response.Result.Rows))
			}

			if !testutil.SliceEql(test.expectedResponse.Columns, response.Result.Columns) {
				t.Fatalf("unexpected column slice, %v got %v", test.expectedResponse.Columns, response.Result.Columns)
			}

			for i := range response.Result.Rows {
				row := response.Result.Rows[i]
				expectedRow := test.expectedResponse.Rows[i]

				if len(*expectedRow) != len(*row) {
					t.Fatalf("unexpected row len, %d got %d", len(*expectedRow), len(*row))
				}

				for ri := range *row {
					field := (*row)[ri]
					expectedField := (*expectedRow)[ri]

					if *expectedField != *field {
						t.Fatalf("unexpected field, %s got %s", *expectedField, *field)
					}
				}
			}
		})
	}
}

type mockRepository struct {}

var _ repository = (*mockRepository)(nil)

func (r *mockRepository) GetAll(ctx context.Context, query string, columns []string) ([]*Row, error) {
	return []*Row{
		{
			new("1"),
			new("Kevin"),
		},
	}, nil
}

func (r *mockRepository) Execute(ctx context.Context, query string) error {
	return nil
}

type mockSchemaService struct {
	tableName		string
	tableSchema	[]string
}

var _ schemaService = (*mockSchemaService)(nil)

func NewMockSchemaService() *mockSchemaService {
	return &mockSchemaService{
		tableName: "users",
		tableSchema: []string{
			"id",
			"name",
			"age",
		},
	}
}

func (m *mockSchemaService) GetSchema(ctx context.Context) (schema.Schema, error) {
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

func (s *mockSchemaService) AppendWildcardSchema(ctx context.Context, tableName string, fields []string) ([]string, error) {
	if tableName != s.tableName {
		return nil, fmt.Errorf("Unknown table, %s got %s", s.tableName, tableName)
	}

	return s.tableSchema, nil
}

type mockDatabaseService struct {}

var _ databaseService = (*mockDatabaseService)(nil)

func NewMockDatabaseService() *mockDatabaseService {
	return &mockDatabaseService{}
}

func (m *mockDatabaseService) GetStats(ctx context.Context) (*database.Stats, error) {
	_, ok := ctx.Value("session_key").(string)
	if !ok {
		return nil, session.ErrNoSessionInContext
	}

	return &database.Stats{
		Version: "0.1.0",
		Size: 100,
	}, nil
}

type mockRegistryService struct {
	registry map[string]int
}

var _ registryService = (*mockRegistryService)(nil)

func NewMockRegistryService() *mockRegistryService {
	return &mockRegistryService{
		registry: map[string]int{
			"users": 1,
			"educations": 2,
		},
	}
}

func (r *mockRegistryService) GetAll(ctx context.Context) (*registry.Registry, error) {
	return &registry.Registry{
		Lookup: r.registry,
	}, nil
}

func (r *mockRegistryService) Create(ctx context.Context, name string) (*int, error) {
	id := len(r.registry)

	r.registry[name] = id

	return &id, nil
}

func (r *mockRegistryService) Update(ctx context.Context, id int, name string) error {
	oldName := r.getName(id)

	if len(oldName) == 0 {
		return registry.ErrInvalidRegistryKey
	}

	delete(r.registry, oldName)
	r.registry[name] = id

	return nil
}

func (r *mockRegistryService) Remove(ctx context.Context, id int) error {
	name := r.getName(id)

	delete(r.registry, name)

	return nil
}

func (r *mockRegistryService) getName(id int) string {
	for name, rId := range r.registry {
		if rId == id {
			return name
		}
	}
	return ""
}
