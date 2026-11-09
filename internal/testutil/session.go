package testutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/fremenkiel/stdout.me/db"
	"github.com/fremenkiel/stdout.me/internal/platform/consts"
)

// Service
type MockSessionService struct {
	Cache	*MockSessionCache
}

func NewMockSessionService(t *testing.T) *MockSessionService {
	return &MockSessionService{
		Cache: NewMockSessionCache(t),
	}
}

func (s *MockSessionService) CreateSessionContext(ctx context.Context) (context.Context, error) {
	id, err := s.CreateSession(ctx)
	if err != nil {
		return nil, err
	}

	return NewContext(ctx, id), nil
}

func (s *MockSessionService) CreateSession(ctx context.Context) (string, error) {
	sessionUuid := uuid.NewV7()
	id := sessionUuid.String()

	if err := s.UpdateSession(id); err != nil {
		return "", err
	}

	return id, nil
}

func (s *MockSessionService) UpdateSession(id string) error {
	if err := s.createSessionDatabase(id); err != nil {
		return err
	}

	return nil
}

func (s *MockSessionService) createSessionDatabase(id string) error {
	databaseName := s.Cache.AddFile(id)
	srcFile, err := db.Files.ReadFile("test.db")
	if err != nil {
		return err
	}

	dstFileHandler, err := os.OpenFile(databaseName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer dstFileHandler.Close()

	if _, err = dstFileHandler.Write(srcFile); err != nil {
		return err
	}

	s.Cache.AddSession(id)

	return nil
}

func (s *MockSessionService) CleanupSession(ctx context.Context) {
	ctxId := ctx.Value(consts.IdKey)

	if ctxId == nil {
		return
	}

	id := ctxId.(string)

	filePath, _ := s.Cache.GetFile(id)

	if err := os.Remove(fmt.Sprintf(filePath, id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}
}

func (s *MockSessionService) GetDatabaseSize(ctx context.Context, id string) (int64, error) {
	if ctx.Err() != nil {
		return -1, ctx.Err()
	}

	fileName, err := s.Cache.GetFile(id)
	if err != nil {
		return -1, err
	}

	file, err := os.Stat(fileName)
	if err != nil {
		return -1, err
	}

	return file.Size(), nil
}

// Cache
type MockSessionCache struct {
	Sessions									sync.Map
	Files											sync.Map
	databasePath,
	databaseFileTemplate			string
}

func NewMockSessionCache(t *testing.T) *MockSessionCache {
	env := LoadTestEnv(t)

	return &MockSessionCache{
			databasePath: env.DatabasePath,
			databaseFileTemplate: env.DatabaseFileTemplate,
	}
}

func (c *MockSessionCache) DatabasePath() string { return c.databasePath }
func (c *MockSessionCache) DatabaseFileTemplate() string { return c.databaseFileTemplate }

func (c *MockSessionCache) LoadSession(id string) (time.Time, bool) {
	session, ok := c.Sessions.Load(id)
	return session.(time.Time), ok
}

func (c *MockSessionCache) AddSession(id string) {
	c.Sessions.Store(id, time.Now())
}

func (c *MockSessionCache) UpdateSession(id string) bool {
	_, loaded := c.Sessions.Swap(id, time.Now())

	return loaded 
}

func (c *MockSessionCache) LoadAndDeleteOldSessions() []string {
	var oldSessions []string

	c.Sessions.Range(func(key any, value any) bool {
		if value.(time.Time).Before(time.Now().Add(-(5*time.Minute))) {
			oldSessions = append(oldSessions, key.(string))
		}
		return true
	})

	for i := range oldSessions {
		c.Sessions.Delete(oldSessions[i])
	}

	return oldSessions
}

func (c *MockSessionCache) LoadAndDeleteOrphantFiles() []string {
	var orphantFiles []string
	c.Files.Range(func(key any, value any) bool {
		id := key.(string)
		if _, ok := c.LoadSession(id); !ok {
			orphantFiles = append(orphantFiles, id)
		}
		return true
	})

	return orphantFiles
}

func (c *MockSessionCache) GetFile(id string) (string, error) {
	str, ok := c.Files.Load(id)
	if !ok {
		return "", os.ErrNotExist
	}

	return str.(string), nil
}

func (c *MockSessionCache) AddFile(id string) string {
	filePath := fmt.Sprintf(c.databaseFileTemplate, id)

	c.Files.Store(id, filePath)

	return filePath
}

func (c *MockSessionCache) RemoveFile(id string) {
	c.Files.Delete(id)
}

func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, consts.IdKey, id)
}

