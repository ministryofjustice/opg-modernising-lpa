# Ship to OPG Metrics Lambda

This lambda is responsible for taking a correctly formatted list of metrics and sends them to the OPG Metrics API Gateway.

## Testing

To debug locally you need to move up to the parent folder and run

`docker-compose up`

This will create a local API that you can curl requests too to test your code.

To view the output of the Lambda you can run

`docker-compose -f 'docker-compose.yml' -p 'lambda' logs -f --tail 1000`
