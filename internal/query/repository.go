package query

import (
	"context"
	"database/sql"

	"github.com/fremenkiel/stdout.me/internal/session"
	_ "modernc.org/sqlite"
)

type sessionCache interface {
	GetFile(id string) (string, error)
}

type Repository struct {
	sessionCache sessionCache
}

var _ repository = (*Repository)(nil)

func NewRepository(sc sessionCache) *Repository {
	return &Repository{sessionCache: sc}
}

func (r *Repository) GetAll(ctx context.Context, query string, columns []string) ([]*Row, error) {
	sessionId, err := session.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	databaseName, err := r.sessionCache.GetFile(sessionId)
	db, err := sql.Open("sqlite", databaseName)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Row

	for rows.Next() {
		rawBuffer := make([]sql.RawBytes, len(columns))
		callArgs := make([]any, len(rawBuffer))

		for i := range rawBuffer {
			callArgs[i] = &rawBuffer[i]
		}

		if err := rows.Scan(callArgs...); err != nil {
			return nil, err
		}

		var row Row
		for i := range rawBuffer {
			str := string(rawBuffer[i])
			row = append(row, &str)
		}

		list = append(list, &row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := db.Close(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Repository) Execute(ctx context.Context, query string) error {
	sessionId, err := session.FromContext(ctx)
	if err != nil {
		return err
	}

	databaseName, err := r.sessionCache.GetFile(sessionId)
	db, err := sql.Open("sqlite", databaseName)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		return err
	}

	return nil
}
