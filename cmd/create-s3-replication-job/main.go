// Create S3 replication job is an AWS Lambda function used to create an S3
// Batch Replication Job to copy files from one S3 bucket to another.
//
// In this service, the source bucket is for uploads to the service and the
// destination bucket is for a case management system in another AWS account.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-lambda-go/otellambda/xrayconfig"
)

var (
	environment = os.Getenv("ENVIRONMENT")
	logger      *slog.Logger
	cfg         aws.Config
)

type configVars struct {
	AccountID                string `json:"aws_account_id"`
	Environment              string `json:'-"`
	ReportAndManifestsBucket string `json:"report_and_manifests_bucket"`
	RoleARN                  string `json:"role_arn"`
	SourceBucket             string `json:"source_bucket"`
}

func main() {
	ctx := context.Background()

	logger = slog.New(telemetry.NewSlogHandler(slog.
		NewJSONHandler(os.Stdout, nil)).
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa/create-s3-replication-job"),
		}))

	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to load default config", slog.Any("err", err))
		return
	}

	tp, err := telemetry.SetupLambda(ctx, &cfg.APIOptions)
	if err != nil {
		logger.WarnContext(ctx, "error creating tracer provider", slog.Any("err", err))
	}

	if tp != nil {
		defer func(ctx context.Context) {
			if err := tp.Shutdown(ctx); err != nil {
				logger.WarnContext(ctx, "error shutting down tracer provider", slog.Any("err", err))
			}
		}(ctx)

		lambda.Start(otellambda.InstrumentHandler(handler, xrayconfig.WithRecommendedOptions(tp)...))
	} else {
		lambda.Start(handler)
	}
}

func handler(ctx context.Context) error {
	vars, err := getVars(ctx, cfg, environment)
	if err != nil {
		return fmt.Errorf("failed to get config vars: %w", err)
	}

	jobID, err := createJob(ctx, cfg, vars)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	logger.InfoContext(ctx, "job created", slog.Any("job_id", jobID))
	return nil
}

func getVars(ctx context.Context, cfg aws.Config, environment string) (configVars, error) {
	ssmClient := ssm.NewFromConfig(cfg)

	param, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String("/modernising-lpa/s3-batch-configuration/" + environment + "/s3_batch_configuration"),
	})
	if err != nil {
		return configVars{}, fmt.Errorf("failed to retrieve parameter: %w", err)
	}

	var vars configVars
	if err := json.Unmarshal([]byte(*param.Parameter.Value), &vars); err != nil {
		return configVars{}, fmt.Errorf("failed to unmarshal parameter: %w", err)
	}

	vars.Environment = environment
	return vars, nil
}

func createJob(ctx context.Context, cfg aws.Config, vars configVars) (string, error) {
	controlClient := s3control.NewFromConfig(cfg)
	requestToken := uuid.NewString()

	resp, err := controlClient.CreateJob(ctx, &s3control.CreateJobInput{
		AccountId:            aws.String(vars.AccountID),
		ConfirmationRequired: aws.Bool(false),
		Operation: &types.JobOperation{
			S3ReplicateObject: &types.S3ReplicateObjectOperation{},
		},
		Report: &types.JobReport{
			Enabled:     true,
			Bucket:      aws.String(vars.ReportAndManifestsBucket),
			Format:      types.JobReportFormatReportCsv20180820,
			ReportScope: types.JobReportScopeAllTasks,
		},
		ClientRequestToken: aws.String(requestToken),
		Description:        aws.String("S3 replication " + vars.Environment + " - golang"),
		Priority:           aws.Int32(10),
		RoleArn:            aws.String(vars.RoleARN),
		ManifestGenerator: &types.JobManifestGeneratorMemberS3JobManifestGenerator{
			Value: types.S3JobManifestGenerator{
				EnableManifestOutput: true,
				ExpectedBucketOwner:  aws.String(vars.AccountID),
				SourceBucket:         aws.String(vars.SourceBucket),
				Filter: &types.JobManifestGeneratorFilter{
					EligibleForReplication:    aws.Bool(true),
					ObjectReplicationStatuses: []types.ReplicationStatus{types.ReplicationStatusFailed, types.ReplicationStatusNone},
				},
				ManifestOutputLocation: &types.S3ManifestOutputLocation{
					ExpectedManifestBucketOwner: aws.String(vars.AccountID),
					Bucket:                      aws.String(vars.ReportAndManifestsBucket),
					ManifestEncryption: &types.GeneratedManifestEncryption{
						SSES3: &types.SSES3Encryption{},
					},
					ManifestFormat: types.GeneratedManifestFormatS3InventoryReportCsv20211130,
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return *resp.JobId, nil
}
