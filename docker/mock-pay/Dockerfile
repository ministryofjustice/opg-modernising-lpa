FROM golang:1.25.7-alpine@sha256:f6751d823c26342f9506c03797d2527668d095b0a15f1862cddb4d927a7a4ced AS base

RUN addgroup -S user && \
  adduser -S -g user user

RUN go install 'github.com/imposter-project/imposter-go/cmd/imposter@v1.8.0' \
  && mv $GOPATH/bin/imposter /tmp/imposter

RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM scratch AS production

ENV IMPOSTER_CONFIG_DIR=/opt/imposter/config
ENV IMPOSTER_PLUGIN_DIR=/opt/imposter/plugins

COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/
COPY --from=base /tmp/imposter /imposter-go
COPY --link ./docker/mock-pay /opt/imposter/config/

USER user
EXPOSE 8080

ENTRYPOINT ["/imposter-go"]
