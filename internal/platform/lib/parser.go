package lib

import (
	"errors"
	"time"

	"github.com/fremenkiel/stdout.me/internal/interpreter"
	"github.com/fremenkiel/stdout.me/internal/ui/viewmodels"
	"modernc.org/sqlite"
	"modernc.org/sqlite/lib"
)

func ParseSqliteError(err error, start time.Time) (*viewmodels.Error, bool) {
	if sqliteError, ok := errors.AsType[*sqlite.Error](err); ok {
		switch sqliteError.Code() {
		case sqlite3.SQLITE_ERROR, 
			sqlite3.SQLITE_ABORT,
			sqlite3.SQLITE_CONSTRAINT,
			sqlite3.SQLITE_INTERRUPT,
			sqlite3.SQLITE_LOCKED,
			sqlite3.SQLITE_PERM,
			sqlite3.SQLITE_RANGE,
			sqlite3.SQLITE_READONLY,
			sqlite3.SQLITE_ROW,
			sqlite3.SQLITE_SCHEMA,
			sqlite3.SQLITE_TOOBIG,
			sqlite3.SQLITE_WARNING:
			return &viewmodels.Error{
				Message: sqliteError.Error(),
				Code: sqlite.ErrorCodeString[sqliteError.Code()],
				Duration: time.Since(start).Microseconds(),
				LineNumber: 0,
				CharStartNumber: 0,
				CharEndNumber: 0,
			}, true
		}
	}

	return nil, false
}

func ParseInterpreterError(err error, start time.Time) (*viewmodels.Error, bool) {
	if interpreterError, ok := errors.AsType[*interpreter.Error](err); ok {
		ln, csn, cen := interpreterError.Pos()
		return &viewmodels.Error{
			Message: interpreterError.Error(),
			Code: interpreter.ErrorCodeString[interpreterError.Code()],
			Duration: time.Since(start).Microseconds(),
			LineNumber: ln,
			CharStartNumber: csn,
			CharEndNumber: cen,
		}, true
	}

	return nil, false
}

func ParseLibError(err error, start time.Time) (*viewmodels.Error, bool) {
	if libError, ok := errors.AsType[*Error](err); ok {
		ln, csn, cen := libError.Pos()
		return &viewmodels.Error{
			Message: libError.Error(),
			Code: ErrorCodeString[libError.Code()],
			Duration: time.Since(start).Microseconds(),
			LineNumber: ln,
			CharStartNumber: csn,
			CharEndNumber: cen,
		}, true
	}

	return nil, false
}
