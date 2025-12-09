FROM node:20-alpine@sha256:d3507a213936fe4ef54760a186e113db5188472d9efdf491686bd94580a1c1e8 AS fetched-repo

WORKDIR /app

RUN apk add --no-cache git=2.49.1-r0
ADD https://api.github.com/repos/ministryofjustice/register-lpa-prototype/git/refs/heads/main version.json
RUN git clone -b main https://github.com/ministryofjustice/register-lpa-prototype.git

FROM node:20-alpine@sha256:d3507a213936fe4ef54760a186e113db5188472d9efdf491686bd94580a1c1e8 AS build

RUN addgroup -g 1017 -S appgroup \
  && adduser -u 1017 -S appuser -G appgroup

COPY scripts/docker_hardening/alpine_image_hardening.sh /harden.sh

WORKDIR /app

COPY --from=fetched-repo app/register-lpa-prototype/package*.json ./

RUN npm install

COPY --from=fetched-repo app/register-lpa-prototype/app ./app
COPY --link ./start.sh ./app/start.sh
RUN chown -R appuser:appgroup /app


RUN /harden.sh && rm /harden.sh


USER 1017

EXPOSE 3000


CMD ["./app/start.sh"]
