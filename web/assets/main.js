import { AwsRum, AwsRumConfig } from 'aws-rum-web';
import * as MOJFrontend from '@ministryofjustice/frontend'
import * as GOVUKFrontend from "govuk-frontend";
import $ from 'jquery';
import { CrossServiceHeader } from './service-header';
import { DataLossWarning } from './data-loss-warning';

window.$ = $

document.addEventListener('DOMContentLoaded', () => {
    GOVUKFrontend.initAll();
    // MOJFrontend.initAll()

    if (typeof MOJFrontend.MultiFileUpload !== "undefined" && document.querySelector('.moj-multi-file-upload')) {
        const csrfToken = document.querySelector("[name='csrf']").getAttribute('value')

        new MOJFrontend.MultiFileUpload({
            container: $(".moj-multi-file-upload"),
            uploadUrl: `${window.location.href}-ajax?csrf=${csrfToken}`,
            deleteUrl: `${window.location.href}-ajax?csrf=${csrfToken}`,
        });
    }

    const header = document.querySelector("[data-module='one-login-header']");
    if (header) {
        new CrossServiceHeader(header).init();
    }

    new DataLossWarning().init()

    const backLink = document.querySelector('.govuk-back-link');
    if (backLink) {
        backLink.addEventListener('click', function (e) {
            window.history.back();
            e.preventDefault();
        }, false);
    }

    function metaContent(name) {
        return document.querySelector(`meta[name=${name}]`).content;
    }

    try {
        const config = {
            sessionSampleRate: 1,
            guestRoleArn: metaContent('rum-guest-role-arn'),
            identityPoolId: metaContent('rum-identity-pool-id'),
            endpoint: metaContent('rum-endpoint'),
            telemetries: ["http", "errors", "performance"],
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
});
