import * as GOVUKFrontend from "govuk-frontend";
import $ from 'jquery'
import { initAll } from '@ministryofjustice/frontend'
import { AwsRum, AwsRumConfig } from 'aws-rum-web';
import { env } from 'node:process';

window.$ = $
initAll()

GOVUKFrontend.initAll();

try {
  const config = {
    sessionSampleRate: 1,
    guestRoleArn: env.AWS_RUM_GUEST_ROLE_ARN,
    identityPoolId: env.AWS_RUM_IDENTITY_POOL_ID,
    endpoint: env.AWS_RUM_ENDPOINT,
    telemetries: ["http","errors","performance"],
    allowCookies: true,
    enableXRay: true
  };

  const APPLICATION_ID = env.AWS_RUM_APPLICATION_ID;
  const APPLICATION_VERSION = '1.0.0';
  const APPLICATION_REGION = env.AWS_RUM_APPLICATION_REGION;

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
