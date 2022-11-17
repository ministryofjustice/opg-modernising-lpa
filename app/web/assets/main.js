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
    identityPoolId: "eu-west-1:7c2ec7b2-7af5-4e45-81da-3c36f3961e1c",
    endpoint: "https://dataplane.rum.eu-west-1.amazonaws.com",
    telemetries: ["http","errors","performance"],
    allowCookies: true,
    enableXRay: true
  };

  const APPLICATION_ID = '3f435cdd-6dec-4cea-837b-33c6103d1562';
  const APPLICATION_VERSION = '1.0.0';
  const APPLICATION_REGION = 'eu-west-1';

  const awsRum = new AwsRum(
    APPLICATION_ID,
    APPLICATION_VERSION,
    APPLICATION_REGION,
    config
  );
} catch (error) {
  console.log(error);
  // Ignore errors thrown during CloudWatch RUM web client initialization
}
