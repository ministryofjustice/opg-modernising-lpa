FROM localstack/localstack:1.3.1 as localstack

COPY ./localstack-init.sh /docker-entrypoint-initaws.d/init.sh
RUN chmod 544 /docker-entrypoint-initaws.d/init.sh
