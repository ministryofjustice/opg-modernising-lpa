import * as GOVUKFrontend from "govuk-frontend";
import $ from 'jquery'
import { initAll } from '@ministryofjustice/frontend'
import { AwsRum, AwsRumConfig } from 'aws-rum-web';

window.$ = $
initAll()

GOVUKFrontend.initAll();

function metaContent(name) {
    return document.querySelector(`meta[name=${name}]`).content;
}

try {
  const config = {
    sessionSampleRate: 1,
    guestRoleArn: metaContent('rum-guest-role-arn'),
    identityPoolId: metaContent('rum-identity-pool-id'),
    endpoint: metaContent('rum-endpoint'),
    telemetries: ["http","errors","performance"],
    allowCookies: true,
    enableXRay: true
  };

  const APPLICATION_ID = metaContent('rum-application-id');
  const APPLICATION_VERSION = '1.0.0';
  const APPLICATION_REGION = metaContent('rum-application-region');

  const awsRum = new AwsRum(
    APPLICATION_ID,
    APPLICATION_VERSION,
    APPLICATION_REGION,
    config
  );
} catch (error) {
  // Ignore errors thrown during CloudWatch RUM web client initialization
}
