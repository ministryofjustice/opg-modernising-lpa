FROM badouralix/curl-jq:latest as build

RUN curl https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json --output openapi-spec.json
RUN cat openapi-spec.json | jq '.paths."/v1/payments/{paymentId}".get.responses."200".content."application/json".examples = { "full-fee": { value: { state: { status: "success" }, amount: 8200, payment_id: "hu20sqlact5260q2nanm0q8u93", reference: "your-reference" } }, "half-fee": { value: { state: { status: "success" }, amount: 4100, payment_id: "hu20sqlact5260q2nanm0q8u93", reference: "your-reference" } }}' > modified-openapi-spec.json

FROM outofcoffee/imposter:latest

COPY --from=build modified-openapi-spec.json /opt/imposter/config/openapi-spec.json
COPY ./docker/mock-pay /opt/imposter/config/
