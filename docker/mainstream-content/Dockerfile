FROM node:24-alpine@sha256:fb71d01345f11b708a3553c66e7c74074f2d506400ea81973343d915cb64eef0 AS fetched-repo

WORKDIR /app

RUN apk add --no-cache git=2.54.0-r0
ADD https://api.github.com/repos/ministryofjustice/register-lpa-prototype/git/refs/heads/main version.json
RUN git clone -b main https://github.com/ministryofjustice/register-lpa-prototype.git

FROM node:24-alpine@sha256:fb71d01345f11b708a3553c66e7c74074f2d506400ea81973343d915cb64eef0 AS production

RUN addgroup -g 1017 -S appgroup \
  && adduser -u 1017 -S appuser -G appgroup

WORKDIR /app

COPY --link docker/mainstream-content/alpine_image_hardening.sh /harden.sh

COPY --from=fetched-repo app/register-lpa-prototype/package*.json ./

# npm 11.13.0 pulls vulnerable tar/undici/brace-expansion transitively; npm 12 includes patched ranges
RUN npm install -g npm@12.0.1

# pin vulnerable transitive deps to patched versions
RUN npm pkg set overrides.uuid=11.1.1 \
  && npm pkg set overrides.tar=7.5.22 \
  && npm pkg set overrides.form-data=4.0.6 \
  && npm pkg set overrides.ws=8.21.1 \
  && npm pkg set overrides.engine.io=6.6.9

RUN npm install

COPY --from=fetched-repo app/register-lpa-prototype/app ./app
COPY --link docker/mainstream-content/start.sh ./app/start.sh
RUN chown -R appuser:appgroup /app


# RUN /harden.sh && rm /harden.sh

USER 1017

EXPOSE 3000

CMD ["./app/start.sh"]
