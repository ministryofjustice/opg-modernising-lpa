FROM golang:1.25.7-alpine@sha256:724e212d86d79b45b7ace725b44ff3b6c2684bfd3131c43d5d60441de151d98e AS base

ARG USERNAME=imposter
ARG USER_UID=1000
ARG USER_GID=1000

RUN addgroup -g $USER_GID $USERNAME \
    && adduser -D -u $USER_UID -G $USERNAME $USERNAME

RUN go install github.com/imposter-project/imposter-go/cmd/imposter@v1.8.0 \
  && cp $GOPATH/bin/imposter /tmp/imposter

RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

FROM scratch AS production

COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/
COPY --from=base /tmp/imposter /imposter-go

EXPOSE 8080
ENTRYPOINT ["/imposter-go"]
