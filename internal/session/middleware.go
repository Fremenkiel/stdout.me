package session

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/fremenkiel/stdout.me/internal/platform/consts"
	"github.com/fremenkiel/stdout.me/internal/platform/logger"
)

type service interface {
	CreateSession(ctx context.Context) (string, error)
	UpdateSession(id string) error
}

func Middleware(s service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sessionId string
			sessionCookie, err := r.Cookie(consts.CookieKey)
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					sessionId, err = s.CreateSession(r.Context())
					if err != nil {
						logger.LogError(
							"session/middlware",
							"error thrown while creating session",
							slog.Any("error", err),
							)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else {
						logger.LogError(
							"session/middlware",
							"error thrown while getting cookie",
							slog.Any("error", err),
							)
					w.WriteHeader(http.StatusUnprocessableEntity)
					return
				}
			} else {
				sessionId = sessionCookie.Value
				if len(sessionId) == 0 {
					sessionId, err = s.CreateSession(r.Context())
					if err != nil {
						logger.LogError(
							"session/middlware",
							"error thrown while creating session",
							slog.Any("error", err),
							)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else {
					if err := s.UpdateSession(sessionId); err != nil {
						logger.LogError(
							"session/middlware",
							"error thrown while updating session",
							slog.Any("error", err),
							)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}

			setCookie(w, sessionId)

			ctx := NewContext(r.Context(), sessionId)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setCookie(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:     consts.CookieKey,
		Value:    id,
		Path:     "/",
		MaxAge:   int((5*time.Minute).Seconds()),
		HttpOnly: true,
	})
}

