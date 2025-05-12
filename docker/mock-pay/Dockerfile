FROM golang:1.24.3-alpine@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044 AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:4.6.8@sha256:3e6e0d67ea6b09b4bedc691ac4e9b7fb1ec6c6309b7abea672d60eeacf83459f

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER imposter

EXPOSE 8080
