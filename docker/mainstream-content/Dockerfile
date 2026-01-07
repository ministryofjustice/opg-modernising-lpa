FROM node:20-alpine@sha256:658d0f63e501824d6c23e06d4bb95c71e7d704537c9d9272f488ac03a370d448 AS fetched-repo

WORKDIR /app

RUN apk add --no-cache git=2.52.0-r0
ADD https://api.github.com/repos/ministryofjustice/register-lpa-prototype/git/refs/heads/main version.json
RUN git clone -b main https://github.com/ministryofjustice/register-lpa-prototype.git

FROM node:20-alpine@sha256:658d0f63e501824d6c23e06d4bb95c71e7d704537c9d9272f488ac03a370d448 AS production

RUN addgroup -g 1017 -S appgroup \
  && adduser -u 1017 -S appuser -G appgroup

WORKDIR /app

COPY --link docker/mainstream-content/alpine_image_hardening.sh /harden.sh

COPY --from=fetched-repo app/register-lpa-prototype/package*.json ./

RUN npm install

COPY --from=fetched-repo app/register-lpa-prototype/app ./app
COPY --link docker/mainstream-content/start.sh ./app/start.sh
RUN chown -R appuser:appgroup /app


# RUN /harden.sh && rm /harden.sh

USER 1017

EXPOSE 3000

CMD ["./app/start.sh"]
