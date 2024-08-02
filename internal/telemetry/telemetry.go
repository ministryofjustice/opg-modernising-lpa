package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/felixge/httpsnoop"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

func Setup(ctx context.Context, resource *resource.Resource) (func(context.Context) error, error) {
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("0.0.0.0:4317"))
	if err != nil {
		return nil, fmt.Errorf("failed to create new OTLP trace exporter: %w", err)
	}

	idg := xray.NewIDGenerator()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	return traceExporter.Shutdown, nil
}

func WrapHandler(handler http.Handler) http.HandlerFunc {
	tracer := otel.GetTracerProvider().Tracer("mlpab")

	return func(w http.ResponseWriter, r *http.Request) {
		route := r.URL.Path
		isWelsh := false
		if strings.HasPrefix(r.URL.Path, "/cy/") {
			route = route[3:]
			isWelsh = true
		}

		target := r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}

		ctx, span := tracer.Start(r.Context(), route,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attribute.Bool("mlpab.welsh", isWelsh)),
			trace.WithAttributes(semconv.HTTPTargetKey.String(target)),
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("mlpab", route, r)...),
		)
		defer span.End()

		m := httpsnoop.CaptureMetrics(handler, w, r.WithContext(ctx))

		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(m.Code)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(m.Code, trace.SpanKindServer))
		span.SetAttributes(semconv.HTTPResponseContentLengthKey.Int64(m.Written))
	}
}

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
