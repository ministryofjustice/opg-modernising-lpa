package telemetry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/smithy-go/middleware"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func AppendMiddlewares(list *[]func(*middleware.Stack) error) {
	*list = append(*list, func(stack *middleware.Stack) error {
		return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("appInitializeMiddlewareAfter",
			func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
				span := trace.SpanFromContext(ctx)

				switch v := in.Parameters.(type) {
				case *eventbridge.PutEventsInput:
					if len(v.Entries) == 1 {
						span.SetAttributes(attribute.String("aws.operation", "PutEvents "+*v.Entries[0].DetailType))
						span.SetAttributes(semconv.CloudeventsEventSource(*v.Entries[0].Source))
						span.SetAttributes(semconv.CloudeventsEventType(*v.Entries[0].DetailType))
					}
				}

				return next.HandleInitialize(ctx, in)
			},
		), middleware.After)
	})
}
