package session

import (
	"context"
	"testing"

	"uuid"

	"github.com/fremenkiel/stdout.me/internal/platform/consts"
)

func TestNewContext(t *testing.T) {
	sessionUuid := uuid.NewV7()
	id := sessionUuid.String()

	ctx := NewContext(t.Context(), id)

	if ctxValue := ctx.Value(consts.IdKey).(string); id != ctxValue {
		t.Fatalf("unexpected context value, %s got %s", id, ctxValue)
	}
}

func TestFromContext(t *testing.T) {
	sessionUuid := uuid.NewV7()
	id := sessionUuid.String()

	ctx := context.WithValue(t.Context(), consts.IdKey, id)

	ctxValue, err := FromContext(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if id != ctxValue {
		t.Fatalf("unexpected context value, %s got %s", id, ctxValue)
	}
}
