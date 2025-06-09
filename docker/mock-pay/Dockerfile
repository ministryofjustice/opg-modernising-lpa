FROM golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:4.6.8@sha256:3e6e0d67ea6b09b4bedc691ac4e9b7fb1ec6c6309b7abea672d60eeacf83459f

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER imposter

EXPOSE 8080
