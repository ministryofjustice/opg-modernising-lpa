# Updating the Lambda Runtime Interface

## Introduction

The Lambda Runtime Interface Emulator (RIE) can be used to locally test a lambda image.

We keep a copy of the RIE in the repository. It can be updated using the following command

```sh
curl -Lo docker/aws-lambda-rie/aws-lambda-rie https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie && \
    chmod +x docker/aws-lambda-rie/aws-lambda-rie
```

## Using the ADOT Collector in a docker container

The following Dockerfile instructions can be used to add the Lambda RIE

```Dockerfile
COPY --link docker/aws-lambda-rie ./aws-lambda-rie
```
