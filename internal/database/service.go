package database

import (
	"context"

	"github.com/fremenkiel/stdout.me/internal/session"
)

type sessionService interface {
	GetDatabaseSize(context.Context, string) (int64, error)
}

type repository interface {
	GetVersion(context.Context) (*string, error)
}

type Service struct {
	repository repository
	sessionService sessionService
}

func NewService(r repository, ss sessionService) *Service {
	return &Service{
		repository: r,
		sessionService: ss,
	}
}

func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	var stats Stats
	version, err := s.repository.GetVersion(ctx)
	if err != nil {
		return nil, err
	}

	stats.Version = *version

	id, err := session.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	size, err := s.sessionService.GetDatabaseSize(ctx, id)
	if err != nil {
		return nil, err
	}

	if size > 0 {
		stats.Size = float64(size) / 1000
	} else {
		stats.Size = -1
	}

	return &stats, nil
}
