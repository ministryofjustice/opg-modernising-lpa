FROM node:24-alpine@sha256:d1b3b4da11eefd5941e7f0b9cf17783fc99d9c6fc34884a665f40a06dbdfc94f AS fetched-repo

WORKDIR /app

RUN apk add --no-cache git=2.52.0-r0
ADD https://api.github.com/repos/ministryofjustice/register-lpa-prototype/git/refs/heads/main version.json
RUN git clone -b main https://github.com/ministryofjustice/register-lpa-prototype.git

FROM node:24-alpine@sha256:d1b3b4da11eefd5941e7f0b9cf17783fc99d9c6fc34884a665f40a06dbdfc94f AS production

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
