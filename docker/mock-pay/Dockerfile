FROM golang:1.24.5-alpine@sha256:daae04ebad0c21149979cd8e9db38f565ecefd8547cf4a591240dc1972cf1399 AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM outofcoffee/imposter:5-beta@sha256:4e5c6ae9b3b5dff725ba67f98b921daad862c65967b4a90021e3812e9bd96f6b
COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

EXPOSE 8080
