FROM golang:1.24.6-alpine@sha256:c8c5f95d64aa79b6547f3b626eb84b16a7ce18a139e3e9ca19a8c078b85ba80d AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

ARG USERNAME=imposter
ARG USER_UID=1000
ARG USER_GID=1000

RUN addgroup -g $USER_GID $USERNAME \
    && adduser -D -u $USER_UID -G $USERNAME $USERNAME

FROM outofcoffee/imposter:5-beta@sha256:c7ab62ecde283e529f0f51b5735f80653789d5fe11a657f7dcf2192431dff970
COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER $USERNAME

EXPOSE 8080
