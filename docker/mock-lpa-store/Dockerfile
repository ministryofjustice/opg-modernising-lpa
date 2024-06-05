FROM alpine:3.19 as build
RUN apk add --no-cache git curl

RUN curl -L https://github.com/outofcoffee/imposter/releases/download/v3.36.0/imposter-plugin-js-graal.zip --output imposter-plugin-js-graal.zip
RUN git clone https://github.com/ministryofjustice/opg-data-lpa-store.git

FROM outofcoffee/imposter:3.40.0

COPY --from=build imposter-plugin-js-graal.zip /opt/imposter/plugins/imposter-plugin-js-graal.zip
COPY ./docker/mock-lpa-store/ /opt/imposter/config/
