FROM golang:1.25.0-alpine@sha256:5dffbd69b37b8e9fc35d94817e94bc7bf4fd047147e3c5353d357bd3492992bf AS base

RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:4.7.0@sha256:2a2f964e6ee7bea2f9a54eac6d976441351d0805833117fc2e45c541b8d58ae9

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER imposter

EXPOSE 8080
