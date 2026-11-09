package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fremenkiel/stdout.me/internal/database"
	"github.com/fremenkiel/stdout.me/internal/platform/logger"
	"github.com/fremenkiel/stdout.me/internal/platform/middleware"
	"github.com/fremenkiel/stdout.me/internal/query"
	"github.com/fremenkiel/stdout.me/internal/registry"
	"github.com/fremenkiel/stdout.me/internal/schema"
	"github.com/fremenkiel/stdout.me/internal/session"
	"github.com/fremenkiel/stdout.me/internal/ui"
	"github.com/fremenkiel/stdout.me/pkg/dotenv"
	uifs "github.com/fremenkiel/stdout.me/ui"
)

func main() {
	logger.NewLogger()

	if err := dotenv.Load(); err != nil {
		logger.LogFatal(
			"main",
			"unable to boot",
			slog.Any("error", err),
			)
	}

	_, err := os.Stat("/tmp/stdout_me_sessions")
	if err != nil {
		if err := os.MkdirAll("/tmp/stdout_me_sessions", os.ModePerm); err != nil {
			logger.LogFatal(
				"main",
				"unable to create tmp folder",
				slog.Any("error", err),
				)
		}
	}

	sessionCache := session.NewCache(
		os.Getenv("DATABASE_PATH"),
		os.Getenv("DATABASE_FILE_TEMPLATE"),
		)

	databaseRepo		:= database.NewRepository(sessionCache)
	queryRepo 			:= query.NewRepository(sessionCache)
	schemaRepo			:= schema.NewRepository(sessionCache)
	registryRepo		:= registry.NewRepository(sessionCache)

	schemaService		:= schema.NewService(schemaRepo)
	sessionService	:= session.NewService(sessionCache)
	databaseService	:= database.NewService(databaseRepo, sessionService)
	registryService	:= registry.NewService(registryRepo)
	rowService			:= query.NewService(queryRepo, schemaService, databaseService, registryService)

	renderer := ui.NewRenderer()

	rowHandler := query.NewHandler(renderer, rowService)

	mux := http.NewServeMux()

	sessionMiddleware := session.Middleware(sessionService)

	staticFiles, err := fs.Sub(uifs.Files, "static")
  if err != nil {
          logger.LogFatal("main", "unable to load static files", slog.Any("error", err))
  }
  mux.Handle("/scripts/", http.FileServer(http.FS(staticFiles)))
  mux.Handle("/styles/", http.FileServer(http.FS(staticFiles)))
  mux.Handle("/images/", http.FileServer(http.FS(staticFiles)))

	requireHtml := middleware.EnforceAccept("text/html")

	query.NewRouter(
		mux,
		rowHandler,
		logger.LogRequestHandler,
		requireHtml,
		sessionMiddleware,
		)

	port := os.Getenv("PORT")
	address := fmt.Sprintf(":%s", port)

	logger.LogInfo(
		"main",
		"server started",
		slog.String("address", "127.0.0.1"),
		slog.String("port", port),
		)

	server := http.Server{
		Addr: address,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sessionService.StartCleanupWorker(ctx, time.Duration(5 * time.Minute))

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.LogInfo(
				"main",
				"http listen and serve error",
				slog.Any("error", err),
				)
		}
	}()

	<-ctx.Done()
	startShutdown := time.Now()
	logger.LogInfo(
		"main",
		"shutdown signel recived",
		slog.Int64("recived_at", startShutdown.Unix()),
		)

	shotdownCtx, shotdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shotdownRelease()

	if err := server.Shutdown(shotdownCtx); err != nil {
		logger.LogInfo(
			"main",
			"unable to gracefully shutdown http server",
			slog.Any("error", err),
			)
	}

	logger.LogInfo(
		"main",
		"shutdown complete",
		slog.Int64("duration_ms", time.Since(startShutdown).Milliseconds()),
		)
}
