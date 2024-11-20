package telemetry

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsClient(t *testing.T) {
	client := newMockCloudwatchClient(t)

	metricsClient := NewMetricsClient(client, "a")

	assert.Equal(t, metricsClient.cloudwatchClient, client)
	assert.Equal(t, metricsClient.baseDimensions[0].Name, aws.String("Version"))
	assert.Equal(t, metricsClient.baseDimensions[0].Value, aws.String("a"))
}

func TestMetricsClientPutMetrics(t *testing.T) {
	ctx := context.Background()
	input := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("namespace"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("a"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(1),
			},
			{
				MetricName: aws.String("b"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(2),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("AnotherDimension"),
						Value: aws.String("your-brain"),
					},
				},
			},
		},
	}

	expectedUpdatedInput := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("namespace"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("a"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(1),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("Version"),
						Value: aws.String("0.0"),
					},
				},
			},
			{
				MetricName: aws.String("b"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(2),
				Dimensions: []types.Dimension{
					{
						Name:  aws.String("AnotherDimension"),
						Value: aws.String("your-brain"),
					},
					{
						Name:  aws.String("Version"),
						Value: aws.String("0.0"),
					},
				},
			},
		},
	}

	client := newMockCloudwatchClient(t)
	client.EXPECT().
		PutMetricData(ctx, expectedUpdatedInput).
		Return(&cloudwatch.PutMetricDataOutput{}, nil)

	metricsClient := NewMetricsClient(client, "0.0")
	err := metricsClient.PutMetrics(ctx, input)

	assert.Nil(t, err)
}

func TestMetricsClientPutMetricsWithoutNamespace(t *testing.T) {
	ctx := context.Background()
	input := &cloudwatch.PutMetricDataInput{
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("a"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(1),
			},
		},
	}

	metricsClient := NewMetricsClient(nil, "0.0")
	err := metricsClient.PutMetrics(ctx, input)

	assert.Nil(t, err)
}
