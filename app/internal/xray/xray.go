package xray

import (
	"context"
	"fmt"
	"log"
	"os"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/ministryofjustice/opg-go-common/logging"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func InitXrayProvider() {
	ctx := context.Background()

	// Create and start new OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint("0.0.0.0:4317"), otlptracegrpc.WithDialOption(grpc.WithBlock()))
	handleErr(err, "failed to create new OTLP trace exporter")

	// Create a new ID Generator
	idg := xray.NewIDGenerator()

	// Instantiate a new ECS Resource detector
	ecsResourceDetector := ecs.NewResourceDetector()
	resource, err := ecsResourceDetector.Detect(context.Background())
	handleErr(err, "failed to instantiate ECS resource detector")

	// Create a new tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithIDGenerator(idg),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	// init aws config
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	handleErr(err, "AWS configuration error initialising XRay provider")

	// instrument all aws clients
	otelaws.AppendMiddlewares(&cfg.APIOptions)
}

func getXrayTraceID(span trace.Span) string {
	xrayTraceID := span.SpanContext().TraceID().String()
	result := fmt.Sprintf("1-%s-%s", xrayTraceID[0:8], xrayTraceID[8:])
	return result
}

func handleErr(err error, message string) {
	logger := logging.New(os.Stdout, "opg-modernising-lpa")
	if err != nil {
		logger.Fatal(err)
		log.Fatal("%s: %v", message, err)
	}
}
