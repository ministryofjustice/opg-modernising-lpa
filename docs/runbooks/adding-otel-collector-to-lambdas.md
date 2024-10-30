# Setting up the AWS Distro for OpenTelemetry Collector (ADOT Collector) for Go Lambda Function Images

## Introduction

The AWS Distro for OpenTelemetry Collector (ADOT Collector) is a distribution of the OpenTelemetry Collector that is optimized for use with AWS services. The ADOT Collector can be used to collect traces from AWS Lambda functions and send them to AWS X-Ray, AWS CloudWatch, or other AWS services.

Instructions for setting up the ADOT Collector for Go Lambda functions using lambda layers are provided in the ADOT Collector documentation: <https://aws-otel.github.io/docs/getting-started/lambda/lambda-go#lambda-layer>, including code examples for using the ADOT Collector with Go Lambda functions.

Layers aren't supported for Image based lambdas, so the ADOT Collector binary needs to be included in the Lambda image. This document provides instructions for including the ADOT Collector binary in a Go Lambda image.

## Downloading the ADOT Collector binary

This method is adapted from <https://github.com/nvsecurity/aws-otel-docker-lambda-layers> and <https://aws.amazon.com/blogs/compute/working-with-lambda-layers-and-extensions-in-container-images/>

The following instructions fetch the URL for the Lambda layer from the official AWS Lambda Layer ARN: <https://aws-otel.github.io/docs/getting-started/lambda/lambda-go>

The URL is used to download the layer zip which is then unzipped to retrieve the ADOT Collector binary.

```sh
export VERSION=0-112-0
URL=$(aws-vault exec management-operator -- aws lambda get-layer-version-by-arn --arn arn:aws:lambda:eu-west-1:901920570463:layer:aws-otel-collector-amd64-ver-${VERSION}:1 --query Content.Location --output text)
curl $URL -o aws-otel-collector-amd64-ver-${VERSION}.zip
unzip aws-otel-collector-amd64-ver-${VERSION}.zip
```

```sh
.
├── aws-otel-collector-amd64-ver-0-102-1.zip
├── collector-config
│   └── config.yaml
└── extensions
    └── collector
```

This method can be used to retrieve the ADOT Collector binary for other languages of the AWS Otel Lambda Layer. The version can be changed by updating the VERSION variable in the script.

## Using the ADOT Collector in a docker container

The following Dockerfile instructions can be used to set up a docker container with the ADOT Collector binary.

```Dockerfile
COPY collector-config/config.yaml /opt/config/config.yaml
COPY extensions/collector /opt/extensions/collector
RUN chmod 755 /opt/config.yaml
ENV AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
ENV OPENTELEMETRY_COLLECTOR_CONFIG_FILE="/opt/config/config.yaml"
```
