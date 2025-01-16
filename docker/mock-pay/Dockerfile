FROM outofcoffee/imposter:4.5.0

COPY ./docker/mock-pay /opt/imposter/config/

ADD https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json /opt/imposter/config/publicapi_spec.json

USER imposter

EXPOSE 8080
