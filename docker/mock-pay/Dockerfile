FROM golang:1.25.7-alpine@sha256:724e212d86d79b45b7ace725b44ff3b6c2684bfd3131c43d5d60441de151d98e AS base

RUN go install 'github.com/imposter-project/imposter-go/cmd/imposter@v1.8.0' \
  && cp $GOPATH/bin/imposter /tmp/imposter

RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM scratch AS production

COPY --from=base /etc/passwd /etc/passwd
COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/
COPY --from=base /tmp/imposter /imposter-go

USER user
EXPOSE 8080

ENTRYPOINT ["/imposter-go"]
