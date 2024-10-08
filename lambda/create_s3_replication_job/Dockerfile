# Define function directory
ARG FUNCTION_DIR="/function"

FROM python:3.12-alpine3.19 AS python-alpine
RUN apk add --no-cache \
    libstdc++=13.2.1_git20231014-r0 \
    elfutils-dev=0.190-r1

FROM python-alpine AS build-image
# Install aws-lambda-cpp build dependencies
RUN apk add --no-cache \
    build-base=0.5-r3 \
    libtool=2.4.7-r3 \
    autoconf=2.71-r2 \
    automake=1.16.5-r2 \
    make=4.4.1-r2 \
    cmake=3.27.8-r0 \
    libcurl=8.9.1-r1

# Include global arg in this stage of the build
ARG FUNCTION_DIR
# Create function directory and set working directory to function root directory
RUN mkdir -p ${FUNCTION_DIR}
WORKDIR ${FUNCTION_DIR}

# Copy function code and requirements.txt
COPY lambda/create_s3_replication_job/src/main.py lambda/create_s3_replication_job/src/requirements.txt ${FUNCTION_DIR}/

# Install the runtime interface client
RUN python -m pip install --root-user-action=ignore \
        --no-cache-dir \
        --target ${FUNCTION_DIR} \
        --requirement requirements.txt

# Multi-stage build: grab a fresh copy of the base image
FROM python-alpine
# Include global arg in this stage of the build
ARG FUNCTION_DIR
# Set working directory to function root directory
WORKDIR ${FUNCTION_DIR}

# Copy in the build image dependencies
COPY --from=build-image ${FUNCTION_DIR} ${FUNCTION_DIR}

ENTRYPOINT [ "/usr/local/bin/python", "-m", "awslambdaric" ]
CMD ["main.handler"]
