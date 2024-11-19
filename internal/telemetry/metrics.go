package telemetry

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type MetricsClient struct {
	cloudwatchClient CloudwatchClient
	baseDimensions   []types.Dimension
}

type CloudwatchClient interface {
	PutMetricData(ctx context.Context, params *cloudwatch.PutMetricDataInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.PutMetricDataOutput, error)
}

func NewMetricsClient(cloudwatchClient CloudwatchClient, versionTag string) *MetricsClient {
	return &MetricsClient{
		cloudwatchClient: cloudwatchClient,
		baseDimensions: []types.Dimension{
			{
				Name:  aws.String("Version"),
				Value: aws.String(versionTag),
			},
		},
	}
}

func (c *MetricsClient) PutMetrics(ctx context.Context, input *cloudwatch.PutMetricDataInput) error {
	if input.Namespace != nil {
		for i, metricDatum := range input.MetricData {
			metricDatum.Dimensions = append(metricDatum.Dimensions, c.baseDimensions...)

			input.MetricData[i] = metricDatum
		}

		_, err := c.cloudwatchClient.PutMetricData(ctx, input)

		return err
	}

	return nil
}
