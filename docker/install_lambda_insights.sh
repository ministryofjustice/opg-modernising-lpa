#!/usr/bin/env bash

TARGETPLATFORM="$1"

if [ "$TARGETPLATFORM" = "linux/arm64" ]; then
  echo "installing for $TARGETPLATFORM"
  curl -O https://lambda-insights-extension-arm64.s3-ap-northeast-1.amazonaws.com/amazon_linux/lambda-insights-extension-arm64.rpm && \
  rpm -U lambda-insights-extension-arm64.rpm && \
  rm -f lambda-insights-extension-arm64.rpm;
fi

if [ "$TARGETPLATFORM" = "linux/amd64" ]; then
echo "installing for $TARGETPLATFORM"
  curl -O https://lambda-insights-extension.s3-ap-northeast-1.amazonaws.com/amazon_linux/lambda-insights-extension.rpm && \
  rpm -U lambda-insights-extension.rpm && \
  rm -f lambda-insights-extension.rpm ;
fi
