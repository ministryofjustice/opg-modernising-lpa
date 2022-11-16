import * as GOVUKFrontend from "govuk-frontend";
import $ from 'jquery'
import { initAll } from '@ministryofjustice/frontend'
import { AwsRum, AwsRumConfig } from 'aws-rum-web';

window.$ = $
initAll()

GOVUKFrontend.initAll();
