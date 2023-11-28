describe('Pay for LPA', () => {
    it('can pay full fee', () => {
        cy.clearCookie('pay');
        cy.getCookie('pay').should('not.exist')

        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('no');

        cy.intercept('**/v1/payments', (req) => {
            cy.getCookie('pay').should('exist');
        });

        cy.contains('button', 'Save and continue').click();

        cy.get('h1').should('contain', 'Payment received');
        cy.checkA11yApp();
        cy.getCookie('pay').should('not.exist');
    });

    it('can apply for a half fee', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('HalfFee');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required to pay a half fee');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="evidence-delivery"]').check('upload');
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/upload-evidence')
        cy.checkA11yApp();

        cy.get('input[type="file"]').attachFile(['dummy.pdf', 'dummy.png']);

        cy.contains('button', 'Upload files').click()

        cy.url().should('contain', '/upload-evidence');

        cy.checkA11yApp();

        cy.get('#dialog').should('not.have.class', 'govuk-!-display-none');
        cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none');
        cy.get('#file-count').should('contain', '0 of 2 files uploaded');

        cy.contains('button', 'Cancel upload').click()
        cy.get('#dialog').should('have.class', 'govuk-!-display-none');
        cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none');

        cy.get('.govuk-summary-list').should('not.exist');

        // spoofing virus scan completing
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&paymentTaskProgress=InProgress&feeType=HalfFee');
        cy.url().should('contain', '/upload-evidence');

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/payment-confirmation');

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();
                cy.visit(`http://localhost:9001/?detail-type=reduced-fee-requested&detail=${uid}`);

                cy.contains('"requestType": "HalfFee"');
                cy.contains(`"evidence": ["${uid}`);
                cy.contains('"evidenceDelivery": "upload"');
            });
    });

    it('can apply for a no fee remission', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('NoFee');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required to pay no fee');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="evidence-delivery"]').check('upload');
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/upload-evidence')
        cy.checkA11yApp();

        cy.get('input[type="file"]').attachFile(['dummy.pdf']);

        cy.contains('button', 'Upload files').click()

        cy.url().should('contain', '/upload-evidence');
        cy.checkA11yApp();

        cy.get('#dialog').should('not.have.class', 'govuk-!-display-none');
        cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none');
        cy.get('#file-count').should('contain', '0 of 1 files uploaded');

        cy.contains('button', 'Cancel upload').click()
        cy.get('#dialog').should('have.class', 'govuk-!-display-none');
        cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none');

        // spoofing virus scan completing
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&paymentTaskProgress=InProgress&feeType=NoFee');
        cy.url().should('contain', '/upload-evidence');

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/evidence-successfully-uploaded');
        cy.checkA11yApp();

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();
                cy.visit(`http://localhost:9001/?detail-type=reduced-fee-requested&detail=${uid}`);

                cy.contains('"requestType": "NoFee"');
                cy.contains(`"evidence": ["${uid}`);
                cy.contains('"evidenceDelivery": "upload"');
            });
    });

    it('can apply for a hardship fee exemption', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('HardshipFee');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required for a hardship application');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="evidence-delivery"]').check('upload');
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/upload-evidence')
        cy.checkA11yApp();

        cy.get('input[type="file"]').attachFile(['dummy.pdf']);

        cy.contains('button', 'Upload files').click()

        cy.url().should('contain', '/upload-evidence');
        cy.checkA11yApp();

        cy.get('#dialog').should('not.have.class', 'govuk-!-display-none');
        cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none');
        cy.get('#file-count').should('contain', '0 of 1 files uploaded');

        cy.contains('button', 'Cancel upload').click()
        cy.get('#dialog').should('have.class', 'govuk-!-display-none');
        cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none');

        // spoofing virus scan completing
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&paymentTaskProgress=InProgress&feeType=NoFee');
        cy.url().should('contain', '/upload-evidence');

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/evidence-successfully-uploaded');
        cy.checkA11yApp();

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();
                cy.visit(`http://localhost:9001/?detail-type=reduced-fee-requested&detail=${uid}`);

                cy.contains('"requestType": "NoFee"');
                cy.contains(`"evidence": ["${uid}`);
                cy.contains('"evidenceDelivery": "upload"');
            });
    });

    it('can only delete evidence that has not been sent to OPG', () => {
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&feeType=HalfFee');
        cy.checkA11yApp();

        cy.url().should('contain', '/upload-evidence');

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png').parent().contains('button span', 'Delete').click();
        });

        cy.url().should('contain', '/upload-evidence');
        cy.checkA11yApp();

        cy.get('.moj-banner').within(() => {
            cy.contains('supporting-evidence.png');
        });
    });

    it('can apply for a reduced fee by posting evidence', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('HardshipFee');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required for a hardship application');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="evidence-delivery"]').check('post');
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/send-us-your-evidence-by-post')
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/what-happens-next-post-evidence');
        cy.checkA11yApp();

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();
                cy.visit(`http://localhost:9001/?detail-type=reduced-fee-requested&detail=${uid}`);

                cy.contains('"requestType": "HardshipFee"');
                cy.contains('"evidence"').should('not.exist');
                cy.contains('"evidenceDelivery": "post"');
            });
    });
});
