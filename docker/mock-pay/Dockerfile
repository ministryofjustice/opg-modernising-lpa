
FROM badouralix/curl-jq:latest as build

RUN curl https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json --output openapi-spec.json
RUN cat openapi-spec.json | jq '.paths."/v1/payments/{paymentId}".get.responses."200".content."application/json".examples = { "full-fee": { value: { amount: 8200 } }, "half-fee": { value: { amount: 4100 } }}' > mktemp && mv mktemp openapi-spec.json

FROM outofcoffee/imposter:latest

COPY --from=build openapi-spec.json /opt/imposter/config/openapi-spec.json
COPY ./docker/mock-pay /opt/imposter/config/
