package query

import (
	"context"

	"github.com/fremenkiel/stdout.me/internal/database"
	"github.com/fremenkiel/stdout.me/internal/interpreter"
	"github.com/fremenkiel/stdout.me/internal/platform/lib"
	"github.com/fremenkiel/stdout.me/internal/registry"
	"github.com/fremenkiel/stdout.me/internal/schema"
)

type repository interface {
	GetAll(ctx context.Context, query string, columns []string) ([]*Row, error)
	Execute(ctx context.Context, query string) error
}

type schemaService interface {
	GetSchema(ctx context.Context) (schema.Schema, error)
	AppendWildcardSchema(ctx context.Context, tableName string, fields []string) ([]string, error)
}

type databaseService interface {
	GetStats(ctx context.Context) (*database.Stats, error)
}

type registryService interface {
	GetAll(ctx context.Context) (*registry.Registry, error)
	Create(ctx context.Context, name string) (*int, error)
	Update(ctx context.Context, id int, name string) error
	Remove(ctx context.Context, id int) error
}

type Service struct {
	repo			repository
	schema		schemaService
	database	databaseService
	registry	registryService
}

var _ service = (*Service)(nil)

func NewService(r repository, ss schemaService, ds databaseService, rs registryService) *Service {
	return &Service{
		repo: r,
		schema: ss,
		database: ds,
		registry: rs,
	}
}

func (s *Service) GetInfo(ctx context.Context) (*Query, error) {
	schema, err := s.getSchema(ctx)
	if err != nil {
		return nil, err
	}

	stats, err := s.getStats(ctx)
	if err != nil {
		return nil, err
	}

	registry, err := s.registry.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return &Query{
		Schema: schema,
		Stats: stats,
		Result: nil, 
		Registry: registry,
	}, nil
}

func (s *Service) Run(ctx context.Context, sql string) (*Query, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var response Query

	parser := interpreter.NewParser()
	trees, err := parser.ParseSql(sql)
	if err != nil {
		return nil, err
	}

	registry, err := s.registry.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var altersDatabase bool
	for i := range trees {
		node := trees[i]

		token := node.Token()
		tableName := node.Table().String()
		returnColumns := node.ReturnColumns()
		switch {
		case i == len(trees) - 1 && returnColumns != nil:
			if tableName == "table_registry" {
				return nil, lib.NewError(
					lib.INVALID_TABLE_NAME,
					token.LineNumber,
					token.End,
					token.Text,
					)
			}
			columns, err := s.schema.AppendWildcardSchema(
				ctx,
				tableName,
				returnColumns,
				)
			if err != nil {
				return nil, err
			}

			var result Result
			result.Columns = columns

			rows, err := s.repo.GetAll(ctx, node.String(), result.Columns)
			if err != nil {
				return nil, err
			}

			result.Rows = rows

			response.Result = &result
			continue
		case token.Type == interpreter.CREATE:
			id, err := s.registry.Create(ctx, tableName)
			if err != nil {
				return nil, err
			}

			registry.Lookup[tableName] = *id

			altersDatabase = true
		case token.Type == interpreter.ALTER:
			alterNode, ok := node.(*interpreter.AlterNode)
			if !ok {
				return nil, lib.NewError(
					lib.INVALID_ALTER_NODE,
					token.LineNumber,
					token.End,
					token.Text,
					)
			}

			newTableName := alterNode.NewTableName()
			if len(newTableName) > 0 {
				id, ok := registry.Lookup[node.Table().String()]
				if !ok {
					return nil, lib.NewError(
						lib.INVALID_TABLE_ALTERATION,
						token.LineNumber,
						token.End,
						token.Text,
						) 
				}

				if err := s.registry.Update(ctx, id, newTableName); err != nil {
					return nil, err
				}

				registry.UpdateName(tableName, newTableName)
			}

			altersDatabase = true
		case token.Type == interpreter.DROP:
			id, ok := registry.Lookup[node.Table().String()]
			if !ok {
				return nil, lib.NewError(
					lib.INVALID_TABLE_ALTERATION,
					token.LineNumber,
					token.End,
					token.Text,
					) 
			}

			if err := s.registry.Remove(ctx, id); err != nil {
				return nil, err
			}

			altersDatabase = true
		case (node.AltersSchema() || node.AltersRows()):
			altersDatabase = true
		}
		if err := s.repo.Execute(ctx, node.String()); err != nil {
			return nil, err
		}
	}

	if altersDatabase {
		schema, err := s.getSchema(ctx)
		if err != nil {
			return nil, err
		}

		response.Schema = schema

		stats, err := s.getStats(ctx)
		if err != nil {
			return nil, err
		}

		response.Stats = stats
	}

	response.Registry = registry

	return &response, nil
}

func (s *Service) getSchema(ctx context.Context) (schema.Schema, error) {
	schema, err := s.schema.GetSchema(ctx)
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func (s *Service) getStats(ctx context.Context) (*database.Stats, error) {
	stats, err := s.database.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
