package session

import (
	"context"
	"errors"

	"github.com/fremenkiel/stdout.me/internal/platform/consts"
)

var ErrNoSessionInContext = errors.New("No session id found in context")

func NewContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, consts.IdKey, id)
}

func FromContext(ctx context.Context) (consts.Id, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	id, ok := ctx.Value(consts.IdKey).(string)

	if !ok {
		return "", ErrNoSessionInContext
	}

	return id, nil
}
