package query

import (
	"github.com/fremenkiel/stdout.me/internal/database"
	"github.com/fremenkiel/stdout.me/internal/registry"
	"github.com/fremenkiel/stdout.me/internal/schema"
)

type Query struct {
	Schema		[]*schema.Table
	Stats			*database.Stats
	Result		*Result
	Registry	*registry.Registry
}

type Result struct {
	Rows			[]*Row
	Columns		[]string
	Duration	int64
}

type Row []*string
