FROM golang:1.24.6-alpine@sha256:c8c5f95d64aa79b6547f3b626eb84b16a7ce18a139e3e9ca19a8c078b85ba80d AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:4.7.0@sha256:2a2f964e6ee7bea2f9a54eac6d976441351d0805833117fc2e45c541b8d58ae9

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER imposter

EXPOSE 8080
