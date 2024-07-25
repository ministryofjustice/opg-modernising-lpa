import { AwsRum, AwsRumConfig } from 'aws-rum-web';
import * as MOJFrontend from '@ministryofjustice/frontend'
import * as GOVUKFrontend from "govuk-frontend";
import $ from 'jquery';
import { CrossServiceHeader } from './service-header';
import { DataLossWarning } from './data-loss-warning';
import { FileUploadModal } from "./file-upload-modal";

// Account for DOMContentLoaded firing before JS runs
if (document.readyState !== "loading") {
    init()
} else {
    document.addEventListener('DOMContentLoaded', init)
}

function init() {
    window.$ = $

    GOVUKFrontend.initAll();
    MOJFrontend.initAll();

    const header = document.querySelector("[data-module='one-login-header']");
    if (header) {
        new CrossServiceHeader(header).init();
    }

    const buttonMenu = document.querySelector(".moj-button-menu");
    if (buttonMenu) {
        new MOJFrontend.ButtonMenu({
            container: buttonMenu,
            mq: "(max-width: 1px)",
            buttonText: "Actions",
            buttonClasses: "govuk-button--secondary moj-button-menu__toggle-button--secondary",
        });
    }

    new DataLossWarning(document.getElementById('return-to-tasklist-btn'), document.getElementById('dialog')).init()
    new DataLossWarning(document.querySelector('.trans-switch a'), document.getElementById('language-dialog')).init()

    new FileUploadModal().init()

    const backLink = document.querySelector('.govuk-back-link');
    if (backLink) {
        backLink.addEventListener('click', function(e) {
            e.preventDefault();

            const previousURL = document.referrer
            const newURL = previousURL + (previousURL.includes('?') ? '&' : '?') + "back=1"
            window.location.replace(newURL)
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
}
