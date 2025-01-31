# Updating the Lambda Runtime Interface

## Introduction

The Lambda Runtime Interface Emulator (RIE) can be used to locally test a lambda image.

We keep a copy of the RIE in the repository. It can be updated using the following command

```sh
curl -Lo docker/aws-lambda-rie/aws-lambda-rie https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/v1.22/download/aws-lambda-rie && \
    chmod +x docker/aws-lambda-rie/aws-lambda-rie
```

## Using the ADOT Collector in a docker container

The following Dockerfile instructions can be used to add the Lambda RIE

```Dockerfile
FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS lambda-rie

WORKDIR /aws-lambda-rie

RUN curl -Lo docker/aws-lambda-rie/aws-lambda-rie https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie && \
chmod +x docker/aws-lambda-rie/aws-lambda-rie

FROM public.ecr.aws/lambda/provided:al2023.2024.10.14.12 AS dev

COPY --from=lambda-rie --link /aws-lambda-rie/aws-lambda-rie ./aws-lambda-rie
```
