package query

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/fremenkiel/stdout.me/internal/database"
	"github.com/fremenkiel/stdout.me/internal/platform/lib"
	"github.com/fremenkiel/stdout.me/internal/platform/logger"
	"github.com/fremenkiel/stdout.me/internal/registry"
	"github.com/fremenkiel/stdout.me/internal/schema"
	"github.com/fremenkiel/stdout.me/internal/ui/viewmodels"
)

type renderer interface {
	RenderPage(w http.ResponseWriter, name string, data any) error
	RenderFragment(w http.ResponseWriter, name string, data any) error
}

type service interface {
	GetInfo(ctx context.Context) (*Query, error)
	Run(ctx context.Context, sql string) (*Query, error)
}

type Handler struct {
	renderer	renderer
	service		service
}

func NewHandler(r renderer, s service) *Handler {
	return &Handler{renderer: r, service: s}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query, err := h.service.GetInfo(ctx)
	if err != nil {
		logger.LogError(
			"table/handler",
			"error returned while getting db info",
			slog.Any("error", err),
			)

		viewData := h.getViewData(nil, &viewmodels.Error{
			Message: "An error returned while getting db info",
			Code: "INTERNAL_SERVER_ERROR",
			Helpertext: "Try to clear the session cookie and refresh the page",
		})
		w.WriteHeader(http.StatusInternalServerError)

		if err := h.renderer.RenderPage(w, "index", viewData); err != nil {
			logger.LogError(
				"table/handler",
				"error returned while rendering",
				slog.Any("error", err),
				)
		}
		return
	} 

	if err := ctx.Err(); err != nil {
		logger.LogError(
			"table/handler",
			"context error",
			slog.Any("error", err),
			)
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}

	viewData := h.getViewData(query, nil)

	if err := h.renderer.RenderPage(w, "index", viewData); err != nil {
		logger.LogError(
			"table/handler",
			"error returned while rendering",
			slog.Any("error", err),
			)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sql := r.URL.Query().Get("sql")
	if len(sql) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	start := time.Now()

	result, err := h.service.Run(ctx, sql)
	if err != nil {
		h.writeError(w, err, start)
		return
	}

	if err := ctx.Err(); err != nil {
		logger.LogError(
			"row/handler",
			"context error",
			slog.Any("error", err),
			)
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}

	if result.Result != nil {
		result.Result.Duration = time.Now().UnixMicro() - start.UnixMicro()
	}

	viewData := h.getViewData(result, nil)

	if err := h.renderer.RenderFragment(w, "result-swap", viewData); err != nil {
		logger.LogError(
			"row/handler",
			"error returned while rendering",
			slog.Any("error", err),
			)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, err error, start time.Time) {
	var libErr *viewmodels.Error

	if parsedErr, ok := lib.ParseInterpreterError(err, start); ok {
		libErr = parsedErr
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else if parsedErr, ok := lib.ParseSqliteError(err, start); ok {
		libErr = parsedErr
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else if parsedErr, ok := lib.ParseLibError(err, start); ok {
		libErr = parsedErr
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	if libErr == nil {
		logger.LogError(
			"row/handler",
			"error returned while getting rows",
			slog.Any("error", err),
			)
		libErr = &viewmodels.Error{
			Message: err.Error(),
			Code: "INTERNAL_SERVER_ERROR",
			Duration: time.Since(start).Microseconds(),
			Helpertext: "Try to clear the session cookie and refresh the page",
		}
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := h.renderer.RenderFragment(w, "error-result-swap", libErr); err != nil {
		logger.LogError(
			"row/handler",
			"error returned while rendering error container",
			slog.Any("error", err),
			)
	}
}

func (h *Handler) getViewData(result *Query, err *viewmodels.Error) *viewmodels.ViewData {
	viewData := viewmodels.ViewData{
		Error: err,
	}

	if result == nil {
		return &viewData
	}

	viewData.Tables = h.getTableViewData(result.Schema, result.Registry)
	viewData.Stats = h.getStatsViewData(result.Stats)
	viewData.Result = h.getResultViewData(result.Result)

	return &viewData
}

func (h *Handler) getTableViewData(tables schema.Schema, registry *registry.Registry) []*viewmodels.Table {
	if tables == nil {
		return nil
	}

	var viewData = make([]*viewmodels.Table, len(tables))

	for i := range tables {
		id, ok := registry.Lookup[tables[i].Name]
		if !ok {
			continue
		}

		viewData[i] = &viewmodels.Table{
			Id: id,
			Name: tables[i].Name,
			RowCount: tables[i].RowCount,
			Columns: h.getColumnViewData(tables[i].Columns),
		}
	}

	return viewData
}

func (h *Handler) getColumnViewData(columns []*schema.Column) *viewmodels.ColumnList {
	viewData := make(viewmodels.ColumnList, len(columns))

	for i := range columns {
		viewData[i] = viewmodels.Column{
			Name: columns[i].Name,
			Type: columns[i].Type,
			IsPrimaryKey: columns[i].IsPrimaryKey,
			IsForeignKey: columns[i].IsForeignKey,
			IsNullable: columns[i].IsNullable,
		}
	}

	return &viewData
}

func (h *Handler) getStatsViewData(stats *database.Stats) *viewmodels.Stats {
	if stats == nil {
		return nil
	}

	version := os.Getenv("VERSION")

	return &viewmodels.Stats{
		Version: version,
		SqliteVersion: stats.Version,
		Size: stats.Size,
	}
}

func (h *Handler) getResultViewData(result *Result) *viewmodels.Result {
	if result == nil {
		return nil
	}

	viewData := &viewmodels.Result{
		Stats: &viewmodels.ResultStats{
			RowCount: uint16(len(result.Rows)),
			ColumnCount: uint16(len(result.Columns)),
			Duration: result.Duration,
		},
		Rows: make([]viewmodels.Row, len(result.Rows)),
		Columns: make(viewmodels.ColumnList, len(result.Columns)),
	}

	for i := range result.Rows {
		viewData.Rows[i] = h.getRowViewModel(result.Rows[i])
	}

	for i, column := range result.Columns {
		viewData.Columns[i].Name = column
	}

	return viewData
}

func (h *Handler) getRowViewModel(row *Row) viewmodels.Row {
	viewModel := make(viewmodels.Row, len(*row))

	for i, val := range *row {
		viewModel[i] = *val
	}

	return viewModel
}
