# Using the Lambda Runtime Interface Emulator

## Introduction

The Lambda Runtime Interface Emulator (RIE) can be used to locally test a lambda image.

These instructions are from [Deploy Go Lambda functions with container images](https://docs.aws.amazon.com/lambda/latest/dg/go-image.html)

The RIE can be downloaded to you home directory with the following command.

```shell
mkdir -p ~/.aws-lambda-rie && \
    curl -Lo ~/.aws-lambda-rie/aws-lambda-rie https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie && \
    chmod +x ~/.aws-lambda-rie/aws-lambda-rie
```

The arm64 emulator is at `https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie-arm64`

## Using the Lambda RIE in a docker container

You can run the container with the RIE by mounting in the volume.

Note the following

- docker-image is the image name and test is the tag.
- /main is the ENTRYPOINT from your Dockerfile.

```shell
docker run --platform linux/amd64 -d -v ~/.aws-lambda-rie:/aws-lambda -p 9000:8080 \
    --entrypoint /aws-lambda/aws-lambda-rie \
    docker-image:test \
        /main
```
