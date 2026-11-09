package schema

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/fremenkiel/stdout.me/internal/session"
	_ "modernc.org/sqlite"
)

type sessionCache interface {
	GetFile(id string) (string, error)
}

type Repository struct {
	cache	sessionCache
}

var _ repository = (*Repository)(nil)

func NewRepository(sc sessionCache) *Repository {
	return &Repository{cache: sc}
}

func (r *Repository) GetSchema(ctx context.Context) (Schema, error) {
	query := `
	SELECT name FROM pragma_table_list WHERE type = 'table' AND name NOT LIKE 'sqlite_%' AND name != 'table_registry' ORDER BY name;
	`

	sessionId, err := session.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	databaseName, err := r.cache.GetFile(sessionId)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", databaseName)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schema Schema

	for rows.Next() {
		var table Table

		if err := rows.Scan(&table.Name); err != nil {
			return nil, err
		}

		countQuery := fmt.Sprintf(`
			SELECT COUNT(*) FROM %s;
			`, table.Name)

		if err := db.QueryRowContext(ctx, countQuery).Scan(&table.RowCount); err != nil {
			return nil, err
		}

		schema = append(schema, &table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := db.Close(); err != nil {
		return nil, err
	}

	return schema, nil
}

func (r *Repository) GetColumns(ctx context.Context, name string) ([]*Column, error) {
	fkQuery := fmt.Sprintf(`
			SELECT "from" FROM pragma_foreign_key_list('%s');
			`, name)
	query := fmt.Sprintf(`
			SELECT * FROM pragma_table_info('%s');
			`, name)

	sessionId, err := session.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	databaseName, err := r.cache.GetFile(sessionId)
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_foreign_keys=on", databaseName))
	if err != nil {
		return nil, err
	}

	fkRows, err := db.QueryContext(ctx, fkQuery)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()

	fkMap := make(map[string]struct{})
	for fkRows.Next() {
		var columnName string
		if err := fkRows.Scan(
			&columnName,
			); err != nil {
			return nil, err
		}

		fkMap[columnName] = struct{}{}
	}

	if err := fkRows.Err(); err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []*Column
	for rows.Next() {
		var column Column
		var defaultValue []byte

		if err := rows.Scan(
			&column.Cid,
			&column.Name,
			&column.Type,
			&column.IsNullable,
			&defaultValue,
			&column.IsPrimaryKey,
			); err != nil {
			return nil, err
		}

		_, ok := fkMap[column.Name]
		column.IsForeignKey = ok

		columns = append(columns, &column)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := db.Close(); err != nil {
		return nil, err
	}

	return columns, nil
}

