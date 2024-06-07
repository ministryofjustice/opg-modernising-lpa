ARG IMPOSTER_VERSION=$IMPOSTER_VERSION

FROM badouralix/curl-jq:latest as build

ARG IMPOSTER_VERSION=$IMPOSTER_VERSION
RUN curl -L "https://github.com/outofcoffee/imposter/releases/download/v${IMPOSTER_VERSION}/imposter-plugin-js-graal.zip" --output imposter-plugin-js-graal.zip

FROM outofcoffee/imposter:${IMPOSTER_VERSION}

COPY --from=build imposter-plugin-js-graal.zip /opt/imposter/plugins/imposter-plugin-js-graal.zip
