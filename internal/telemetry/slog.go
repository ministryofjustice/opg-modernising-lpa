package telemetry

import (
	"context"
	"log/slog"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"go.opentelemetry.io/otel/trace"
)

type SlogHandler struct {
	handler slog.Handler
}

func NewSlogHandler(h slog.Handler) slog.Handler {
	return &SlogHandler{handler: h}
}

func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *SlogHandler) Handle(ctx context.Context, record slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID()
		record.AddAttrs(slog.String("trace_id", traceID.String()))
	}

	session, err := appcontext.SessionFromContext(ctx)
	if err == nil {
		record.AddAttrs(slog.String("session_id", session.SessionID))

		if session.OrganisationID != "" {
			record.AddAttrs(slog.String("organisation_id", session.OrganisationID))
		}
	}

	return h.handler.Handle(ctx, record)
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewSlogHandler(h.handler.WithAttrs(attrs))
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	return NewSlogHandler(h.handler.WithGroup(name))
}
