package session

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/fremenkiel/stdout.me/internal/testutil"
)

func TestAddSession(t *testing.T) {
	tests := []struct{
		name				string
		id					string
	} {
		{
			name: "add_session",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func (t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				env := testutil.LoadTestEnv(t)
				cache := NewCache(env.DatabasePath, env.DatabaseFileTemplate)

				cache.AddSession(test.id)

				cache.sessions.Range(func(key, value any) bool {
					id := key.(string)
					addTime := value.(time.Time)

					if test.id != id {
						t.Fatalf("unexpected session key, %s got %s",test.id, id)
					}

					if !time.Now().Equal(addTime) {
						t.Fatalf("unexpected add time, %s got %s", time.Now().String(), addTime.String())
					}

					return true
				})
			})
		})
	}
}

func TestUpdateSession(t *testing.T) {
	tests := []struct{
		name							string
		id								string
		loaded						bool
		expectedResponse	bool
	} {
		{
			name: "update_session",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: true,
			expectedResponse: true,
		},
		{
			name: "not_loaded",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: false,
			expectedResponse: false,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func (t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				env := testutil.LoadTestEnv(t)
				cache := &Cache{
					databasePath: env.DatabasePath,
					databaseFileTemplate: env.DatabaseFileTemplate,
				}

				if test.loaded {
					cache.sessions.Store(test.id, time.UTC)
				}

				response := cache.UpdateSession(test.id)

				if test.expectedResponse != response {
					t.Fatalf("unexpected response, %v got %v", test.expectedResponse, response)
				}

				cache.sessions.Range(func(key, value any) bool {
					id := key.(string)
					addTime := value.(time.Time)

					if test.id != id {
						t.Fatalf("unexpected session key, %s got %s",test.id, id)
					}

					if !time.Now().Equal(addTime) {
						t.Fatalf("unexpected add time, %s got %s", time.Now().String(), addTime.String())
					}

					return true
				})
			})
		})
	}
}

func TestLoadAndDelete(t *testing.T) {
	tests := []struct {
		name						string
		initialState		map[string]int64
		expectedResponse	[]string
	} {
		{
			name: "two_expired_entries",
			initialState: map[string]int64{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 4,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 10,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 60,
			},
			expectedResponse: []string{
				"01a100bf-c3be-7dc7-abe5-f68536f492a6",
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1",
			},
		},
		{
			name: "no_expired_entries",
			initialState: map[string]int64{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 0,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 2,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 3,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 4,
			},
			expectedResponse: []string{},
		},
		{
			name: "all_expired_entries",
			initialState: map[string]int64{
				"1a100bf-c3be-7d9c-acb8-785e8722f892": 5,
				"01a100bf-c3be-7dbb-b038-49c4e6479dba": 6,
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1": 60,
				"01a100bf-c3be-7dc7-abe5-f68536f492a6": 120,
			},
			expectedResponse: []string{
				"1a100bf-c3be-7d9c-acb8-785e8722f892",
				"01a100bf-c3be-7dbb-b038-49c4e6479dba",
				"01a100bf-c3be-7dc3-ac37-5ac1e37cc6b1",
				"01a100bf-c3be-7dc7-abe5-f68536f492a6",
			},
		},
		{
			name: "no_entries",
			initialState: map[string]int64{},
			expectedResponse: []string{},
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func(t *testing.T) {
			env := testutil.LoadTestEnv(t)
			cache := Cache{
				databasePath: env.DatabasePath,
				databaseFileTemplate: env.DatabaseFileTemplate,
			}

			for key, value := range test.initialState {
				lastSeen := time.Now().Add(-(time.Duration(value)*time.Minute))
				cache.sessions.Store(key, lastSeen)
			}

			response := cache.LoadAndDeleteOldSessions()

			if !testutil.SliceEql(test.expectedResponse, response) {
				t.Fatalf("response do not match, expected %v, got %v", test.expectedResponse, response)
			}

			remainder := make([]string, len(test.initialState) - len(test.expectedResponse))
			for key := range test.initialState {
				if !slices.Contains(test.expectedResponse, key) {
					remainder = append(remainder, key)
				}
			}

			cache.sessions.Range(func(key, value any) bool {
				id := key.(string)
				if !slices.Contains(remainder, id) {
					t.Fatalf("unexpected session present in cache, got %s", id)
				}
				return true
			})
		})
	}
}

func TestAddFile(t *testing.T) {
	tests := []struct{
		name							string
		id								string
		loaded						bool
	} {
		{
			name: "add_file",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: true,
		},
		{
			name: "not_loaded",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: false,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func (t *testing.T) {
			env := testutil.LoadTestEnv(t)
			fileName := fmt.Sprintf(env.DatabaseFileTemplate, test.id)
			cache := &Cache{
				databasePath: env.DatabasePath,
				databaseFileTemplate: env.DatabaseFileTemplate,
			}

			if test.loaded {
				cache.files.Store(test.id, fileName)
			}

			response := cache.AddFile(test.id)

			path, ok := cache.files.Load(test.id)

			if !ok {
				t.Fatal("file was not loaded to cache")
			}

			if !strings.Contains(path.(string), fileName) {
				t.Fatalf("unexpected file path, %s got %s", fileName, path)
			}

			if !strings.Contains(response, fileName) {
				t.Fatalf("unexpected response, %s got %s", fileName, response)
			}
		})
	}
}

func TestRemoveFile(t *testing.T) {
	tests := []struct{
		name							string
		id								string
		loaded						bool
	} {
		{
			name: "remove_file",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: true,
		},
		{
			name: "not_loaded",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: false,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func (t *testing.T) {
			env := testutil.LoadTestEnv(t)
			fileName := fmt.Sprintf(env.DatabaseFileTemplate, test.id)
			cache := &Cache{
				databasePath: env.DatabasePath,
				databaseFileTemplate: env.DatabaseFileTemplate,
			}

			if test.loaded {
				cache.files.Store(test.id, fileName)
			}

			cache.RemoveFile(test.id)

			if _, ok := cache.files.Load(test.id); ok {
				t.Fatal("file was not removed from cache")
			}
		})
	}
}

func TestGetFile(t *testing.T) {
	tests := []struct{
		name											string
		id												string
		loaded										bool
		expectedResponseContains	string
		expectedError							error
	} {
		{
			name: "remove_file",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: true,
			expectedResponseContains: "1a100bf-c3be-7d9c-acb8-785e8722f892.db",
			expectedError: nil,
		},
		{
			name: "not_loaded",
			id: "1a100bf-c3be-7d9c-acb8-785e8722f892",
			loaded: false,
			expectedResponseContains: "",
			expectedError: os.ErrNotExist,
		},
	}

	for ti := range tests {
		test := tests[ti]

		t.Run(test.name, func (t *testing.T) {
			env := testutil.LoadTestEnv(t)
			fileName := fmt.Sprintf(env.DatabaseFileTemplate, test.id)
			cache := &Cache{
				databasePath: env.DatabasePath,
				databaseFileTemplate: env.DatabaseFileTemplate,
			}

			if test.loaded {
				cache.files.Store(test.id, fileName)
			}

			path, err := cache.GetFile(test.id)

			if test.expectedError != err {
				t.Fatalf("unexpected error, %v got %v", test.expectedError, err)
			}

			if !strings.Contains(path, test.expectedResponseContains) {
				t.Fatalf("unexpected path, %s got %s", test.expectedResponseContains, path)
			}
		})
	}
}
