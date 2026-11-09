package database

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
	sessionCache	sessionCache
}

var _ repository = (*Repository)(nil)

func NewRepository(sc sessionCache) *Repository {
	return &Repository{sessionCache: sc}
}

func (r *Repository) GetVersion(ctx context.Context) (*string, error) {
	query := `
	SELECT sqlite_version();
	`

	sessionId, err := session.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	databaseName, err := r.sessionCache.GetFile(sessionId)
	db, err := sql.Open("sqlite", databaseName)
	if err != nil {
		return nil, err
	}

	var version string

	if err := db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		return nil, err
	}

	if err := db.Close(); err != nil {
		return nil, err
	}

	return &version, nil
}
