// Package telemetry provides functionality for tracing with AWS X-Ray and
// logging information related to the current web request.
package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/smithy-go/middleware"
	"github.com/felixge/httpsnoop"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

func Setup(ctx context.Context, resource *resource.Resource, apiOptions *[]func(*middleware.Stack) error) (func(context.Context) error, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("0.0.0.0:4317"))
	if err != nil {
		return nil, fmt.Errorf("failed to create new OTLP trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	idg := xray.NewIDGenerator()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithIDGenerator(idg),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	otelaws.AppendMiddlewares(apiOptions)
	AppendMiddlewares(apiOptions)

	return tp.Shutdown, nil
}

// SetupLambda requires an ADOT collector lambda extension to be included in the
// deployed lambda image for the below instrumentation to work
func SetupLambda(ctx context.Context, apiOptions *[]func(*middleware.Stack) error) (*sdktrace.TracerProvider, error) {
	tp, err := xrayconfig.NewTracerProvider(ctx)
	if err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	otelaws.AppendMiddlewares(apiOptions)
	AppendMiddlewares(apiOptions)

	return tp, nil
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
