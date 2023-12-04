FROM localstack/localstack:3.0.2 as localstack

COPY localstack-init.sh /etc/localstack/init/ready.d/localstack-init.sh
COPY dynamodb-lpa-gsi-schema.json ./dynamodb-lpa-gsi-schema.json

RUN chmod 544 /etc/localstack/init/ready.d/localstack-init.sh
