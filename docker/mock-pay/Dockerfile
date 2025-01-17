FROM golang:1.23.4-alpine AS base
RUN apk add --no-cache wget
WORKDIR /opt
RUN wget -q https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:4.5.0

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /opt/publicapi_spec.json /opt/imposter/config/

WORKDIR /opt/imposter/config/


USER imposter

EXPOSE 8080
