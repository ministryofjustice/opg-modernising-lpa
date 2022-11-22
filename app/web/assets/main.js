import * as GOVUKFrontend from "govuk-frontend";
import $ from 'jquery'
import { initAll } from '@ministryofjustice/frontend'
import { AwsRum, AwsRumConfig } from 'aws-rum-web';

window.$ = $
initAll()

GOVUKFrontend.initAll();

try {
  const config = {
    sessionSampleRate: 1,
    guestRoleArn: "arn:aws:iam::653761790766:role/RUM-Monitor-Unauthenticated",
    identityPoolId: "eu-west-1:1b1631fd-d258-4677-a11d-51096cc876c1",
    endpoint: "https://dataplane.rum.eu-west-1.amazonaws.com",
    telemetries: ["http","errors","performance"],
    allowCookies: true,
    enableXRay: true
  };

  const APPLICATION_ID = '81df3db5-de4f-46b1-869e-60cf0e4ac165';
  const APPLICATION_VERSION = '1.0.0';
  const APPLICATION_REGION = 'eu-west-1';

  const awsRum = new AwsRum(
    APPLICATION_ID,
    APPLICATION_VERSION,
    APPLICATION_REGION,
    config
  );
} catch (error) {
  // Ignore errors thrown during CloudWatch RUM web client initialization
}
