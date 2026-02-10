FROM golang:1.25.7-alpine@sha256:724e212d86d79b45b7ace725b44ff3b6c2684bfd3131c43d5d60441de151d98e AS base
RUN wget -q -O /tmp/publicapi_spec.json https://raw.githubusercontent.com/alphagov/pay-publicapi/master/openapi/publicapi_spec.json

ARG USERNAME=imposter
ARG USER_UID=1000
ARG USER_GID=1000

RUN addgroup -g $USER_GID $USERNAME \
    && adduser -D -u $USER_UID -G $USERNAME $USERNAME

FROM outofcoffee/imposter:5-beta@sha256:c85f7e7c3434cda5fdf8b7100a98b91ec073e8cb6500367560b24c58c6faa264 AS production
COPY ./docker/mock-pay /opt/imposter/config/

COPY --from=base /tmp/publicapi_spec.json /opt/imposter/config/

USER $USERNAME

EXPOSE 8080
