FROM golang:1.24.1-alpine@sha256:43c094ad24b6ac0546c62193baeb3e6e49ce14d3250845d166c77c25f64b0386 AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:4.6.3@sha256:fc1e3b9cbef1ec435f292c9b5f1e4da566ae7e89fa5e766d9eca9c31cc8c9717

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER imposter

EXPOSE 8080
