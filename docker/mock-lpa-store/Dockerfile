FROM badouralix/curl-jq:latest as build

RUN curl -L https://github.com/outofcoffee/imposter/releases/download/v3.36.0/imposter-plugin-js-graal.zip --output imposter-plugin-js-graal.zip

FROM outofcoffee/imposter:3.40.0

COPY --from=build imposter-plugin-js-graal.zip /opt/imposter/plugins/imposter-plugin-js-graal.zip
