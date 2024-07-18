FROM badouralix/curl-jq:latest as build

RUN curl https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json --output openapi-spec.json

FROM outofcoffee/imposter:latest

COPY --from=build openapi-spec.json /opt/imposter/config/openapi-spec.json
COPY ./docker/mock-pay /opt/imposter/config/
