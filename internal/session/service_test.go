package session

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fremenkiel/stdout.me/internal/testutil"
	"uuid"
)

func TestCreateSession(t *testing.T) {
	cache := testutil.NewMockSessionCache(t)
	service := NewService(cache)

	ctx := t.Context()
	id, err := service.CreateSession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ctx = NewContext(ctx, id)
	defer service.removeSessionDatabase(id)

	if _, ok := cache.Sessions.Load(id); !ok {
		t.Fatal("session not saved in cache")
	}

	if _, ok := cache.Files.Load(id); !ok {
		t.Fatal("session database file path not saved in cache")
	}
}

func TestServiceUpdateSession(t *testing.T) {
	tests := []struct{
		name			string
		loaded		bool
	} {
		{
			name: "update_existing",
			loaded: true,
		},
		{
			name: "create_missing",
			loaded: false,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			cache := testutil.NewMockSessionCache(t)
			service := NewService(cache)

			sessionUuid := uuid.NewV7()
			id := sessionUuid.String()

			if test.loaded {
				cache.Sessions.Store(id, nil)
				cache.Files.Store(id, "placeholder_value")
			}

			if err := service.UpdateSession(id); err != nil {
				t.Fatal(err)
			}
			defer service.removeSessionDatabase(id)

			if _, ok := cache.Sessions.Load(id); !ok {
				t.Fatal("session not saved in cache")
			}

			if _, ok := cache.Files.Load(id); !ok {
				t.Fatal("session database file path not saved in cache")
			}
		})
	}
}

func TestStartCleanupWorker(t *testing.T) {
	tests := []struct{
		name					string
		initialState	string
		expectedError	error
	} {
		{
			name: "remove_session",
			initialState: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			expectedError: nil,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				env := testutil.LoadTestEnv(t)
				cache := NewCache(env.DatabasePath, env.DatabaseFileTemplate)
				service := NewService(cache)

				lastSeen := time.Now().Add(-(5*time.Minute))
				cache.sessions.Store(test.initialState, lastSeen)

				fileName := fmt.Sprintf(env.DatabaseFileTemplate, test.initialState)
				cache.sessions.Store(test.initialState, time.Now().Add(-(6*time.Minute)))
				cache.files.Store(test.initialState, fileName)

				if err := createTestFile(fileName); err != nil {
					t.Fatal(err)
				}

				ctx := t.Context()
				service.StartCleanupWorker(ctx, 5 * time.Second)

				synctest.Sleep(5 * time.Second)

				if _, err := os.Stat(fileName); err == nil {
					t.Fatal("file was not removed")
				}

				if _, ok := cache.sessions.Load(test.initialState); ok {
					t.Fatalf("session not removed from cache, id %s", test.initialState)
				}

				if _, ok := cache.files.Load(test.initialState); ok {
					t.Fatalf("session file path not removed from cache, id %s", test.initialState)
				}
			})
		})
	}
}

func TestRemoveExpiredSessions(t *testing.T) {
	tests := []struct{
		name					string
		initialState			map[string]int16	
		sessionLoaded	bool
		fileLoaded		bool
		fileCreated		bool
		expectedError	error
	} {
		{
			name: "remove_session",
			initialState: map[string]int16{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 4,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 2,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 1,
			},
			sessionLoaded: true,
			fileLoaded: true,
			fileCreated: true,
			expectedError: nil,
		},
		{
			name: "session_not_loaded",
			initialState: map[string]int16{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 4,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 2,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 1,
			},
			sessionLoaded: false,
			fileLoaded: true,
			fileCreated: true,
			expectedError: nil,
		},
		{
			name: "file_not_created",
			initialState: map[string]int16{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 4,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 2,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 1,
			},
			sessionLoaded: true,
			fileLoaded: true,
			fileCreated: false,
			expectedError: nil,
		},
		{
			name: "file_not_loaded",
			initialState: map[string]int16{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 4,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 2,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 1,
			},
			sessionLoaded: true,
			fileLoaded: false,
			fileCreated: true,
			expectedError: nil,
		},
		{
			name: "file_not_loaded_remove_all",
			initialState: map[string]int16{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 4,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 2,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 10,
			},
			sessionLoaded: true,
			fileLoaded: false,
			fileCreated: true,
			expectedError: nil,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			env := testutil.LoadTestEnv(t)
			cache := NewCache(env.DatabasePath, env.DatabaseFileTemplate)
			service := NewService(cache)

			var expired []string
			
			for key, value := range test.initialState {
				lastSeen := time.Now().Add(-(time.Duration(value)*time.Minute))
				cache.sessions.Store(key, lastSeen)
				if value > 5 {
					expired = append(expired, key)
				}
			}

			sessionUuid := uuid.NewV7()
			id := sessionUuid.String()
			fileName := fmt.Sprintf(env.DatabaseFileTemplate, id)

			expired = append(expired, id)

			if test.sessionLoaded {
				cache.sessions.Store(id, time.Now().Add(-(6*time.Minute)))
			}

			if test.fileLoaded {
				cache.files.Store(id, fileName)
			}

			if test.fileCreated {
				if err := createTestFile(fileName); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := ensureFolder(); err != nil {
					t.Fatal(err)
				}
			}

			if err := service.RemoveExpiredSessions(); !errors.Is(err, test.expectedError) {
				t.Fatalf("unexpected error, %v got %v", test.expectedError, err)
			}

			if _, err := os.Stat(fileName); err == nil {
				t.Fatal("file was not removed")
			}

			for _, session := range expired {
				if _, ok := cache.sessions.Load(session); ok {
					t.Fatalf("session not removed from cache, id %s", session)
				}

				if _, ok := cache.files.Load(session); ok {
					t.Fatalf("session file path not removed from cache, id %s", session)
				}
			}
		})
	}
}

func TestCreateSessionDatabase(t *testing.T) {
	cache := testutil.NewMockSessionCache(t)
	service := NewService(cache)

	if err := ensureFolder(); err != nil {
		t.Fatal(err)
	}

	id := uuid.NewV7()
	if err := service.createSessionDatabase(id.String()); err != nil {
		t.Fatal(err)
	}
	defer service.removeSessionDatabase(id.String())

	count := 0
	cache.Files.Range(func(key, value any) bool {
		if count > 1 {
			t.Fatal("unexpected file len, exceded 1")
		}
		count++
		return true
	})

	databaseName, err := cache.GetFile(id.String())
	if err != nil {
		t.Fatalf("unexpected error while getting file, %v", err)
	}

	if _, err := os.Stat(databaseName); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveSessionDatabase(t *testing.T) {
	env := testutil.LoadTestEnv(t)
	cache := NewCache(env.DatabasePath, env.DatabaseFileTemplate)
	service := NewService(cache)

	id := uuid.NewV7()
	databaseName := cache.AddFile(id.String())

	if err := createTestFile(databaseName); err != nil {
		t.Fatal(err)
	}

	if err := service.removeSessionDatabase(id.String()); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(databaseName); err == nil {
		t.Fatal("file was not removed")
	}

	count := 0
	cache.files.Range(func(key, value any) bool {
		if count > 0 {
			t.Fatal("unexpected file len, exceded 0")
		}
		count++
		return true
	})
}

func createTestFile(name string) error {
	if err := ensureFolder(); err != nil {
		return err
	}

	dstFileHandler, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer dstFileHandler.Close()

	if _, err = dstFileHandler.Write([]byte{1}); err != nil {
		return err
	}

	return nil
}

func ensureFolder() error {
	_, err := os.Stat("/tmp/stdout_me_sessions")
	if err != nil {
		if err := os.MkdirAll("/tmp/stdout_me_sessions", os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

