describe('Pay for LPA', { pageLoadTimeout: 8000 }, () => {
    it('can pay full fee', () => {
        cy.clearCookie('pay');
        cy.getCookie('pay').should('not.exist')

        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('no', { force: true });

        cy.intercept('**/v1/payments', (req) => {
            cy.getCookie('pay').should('exist');
        });

        cy.contains('button', 'Save and continue').click();

        cy.get('h1').should('contain', 'Payment received');
        cy.checkA11yApp();
        cy.getCookie('pay').should('not.exist');

        cy.contains('a', 'Continue').click();
        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=payment-received&detail=${uid}`);
                    cy.contains('"amount":8200');
                    cy.contains('"paymentId":"hu20sqlact5260q2nanm0q8u93"');
                });
            });
    });

    it('can apply for a half fee', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('HalfFee', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required to pay a half fee');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="selected"]').check('upload', { force: true });
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

        cy.get('#uploaded .govuk-summary-list').should('not.exist');

        // spoofing virus scan completing
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&paymentTaskProgress=InProgress&feeType=HalfFee');
        cy.url().should('contain', '/upload-evidence');

        cy.get('form#delete-form .govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/payment-successful');
        cy.contains('a', 'Continue').click()

        cy.url().should('contain', '/evidence-successfully-uploaded');
        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', "Pay for the LPA").should('contain', 'Pending').click();

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();
        cy.contains('Half fee (remission)');

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();

                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=reduced-fee-requested&detail=${uid}`);
                    cy.contains('"requestType":"HalfFee"');
                    cy.contains(new RegExp(`{"path":"${uid}/evidence/.+","filename":"supporting-evidence.png"}`))
                    cy.contains('"evidenceDelivery":"upload"');
                });
            });
    });

    it('can apply for a no fee exemption', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('NoFee', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required to pay no fee');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="selected"]').check('upload', { force: true });
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

        cy.get('#uploaded .govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()
        cy.url().should('contain', '/task-list');
        cy.contains('li', "Pay for the LPA").should('contain', 'Pending').click();

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();
        cy.contains('No fee (exemption)');

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();

                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=reduced-fee-requested&detail=${uid}`);
                    cy.contains('"requestType":"NoFee"');
                    cy.contains(new RegExp(`{"path":"${uid}/evidence/.+","filename":"supporting-evidence.png"}`))
                    cy.contains('"evidenceDelivery":"upload"');
                });
            });
    });

    it('can apply for a hardship fee exemption', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('HardshipFee', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required for a hardship application');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="selected"]').check('upload', { force: true });
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
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&paymentTaskProgress=InProgress&feeType=HardshipFee');
        cy.url().should('contain', '/upload-evidence');

        cy.get('#uploaded .govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()
        cy.url().should('contain', '/task-list');
        cy.contains('li', "Pay for the LPA").should('contain', 'Pending').click();

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();
        cy.contains('Hardship');

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();

                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=reduced-fee-requested&detail=${uid}`);
                    cy.contains('"requestType":"HardshipFee"');
                    cy.contains(new RegExp(`{"path":"${uid}/evidence/.+","filename":"supporting-evidence.png"}`))
                    cy.contains('"evidenceDelivery":"upload"');
                });
            });
    });

    it('can delete evidence that has not been sent to OPG', () => {
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&feeType=HalfFee');
        cy.checkA11yApp();

        cy.url().should('contain', '/upload-evidence');

        cy.get('#uploaded .govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png').parent().contains('button', 'Delete').click();
        });

        cy.url().should('contain', '/upload-evidence');
        cy.checkA11yApp();

        cy.get('.govuk-notification-banner').within(() => {
            cy.contains('supporting-evidence.png');
        });
    });

    it('can see evidence that has previously been sent to OPG', () => {
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&feeType=HalfFee');
        cy.checkA11yApp();

        cy.url().should('contain', '/upload-evidence');

        cy.contains('a', 'Previously uploaded files').click()

        cy.get('#previouslyUploaded').within(() => {
            cy.contains('previously-uploaded-evidence.png');
        });
    });

    it('can apply for a reduced fee by posting evidence', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('HardshipFee', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Evidence required for a hardship application');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="selected"]').check('post', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/send-us-your-evidence-by-post')
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();

        cy.visit('/dashboard');
        cy.contains('.govuk-body-s', 'Reference number:')
            .invoke('text')
            .then((text) => {
                const uid = text.split(':')[1].trim();

                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=reduced-fee-requested&detail=${uid}`);
                    cy.contains('"requestType":"HardshipFee"');
                    cy.contains('"evidence"').should('not.exist');
                    cy.contains('"evidenceDelivery":"post"');
                });
            });
    });

    it('can pay remaining amount when approved', () => {
        cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa&paymentTaskProgress=Approved&feeType=HalfFee');

        cy.intercept('**/v1/payments', (req) => {
            cy.getCookie('pay').should('exist');
        });

        cy.contains('li', 'Pay for the LPA').should('contain', 'In progress').click();

        cy.get('h1').should('contain', 'Payment received');
        cy.checkA11yApp();
        cy.getCookie('pay').should('not.exist');

        cy.contains('a', 'Continue').click();

        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=payment-received&detail=${uid}`);
                    cy.contains('"amount":4100');
                    cy.contains('"paymentId":"hu20sqlact5260q2nanm0q8u93"');
                });
            });
    });

    it('can pay remaining amount when denied', () => {
        cy.visit('/fixtures?redirect=/task-list&progress=payForTheLpa&paymentTaskProgress=Denied&feeType=HalfFee');

        cy.contains('li', 'Pay for the LPA').should('contain', 'Denied').click();

        cy.intercept('**/v1/payments', (req) => {
            cy.getCookie('pay').should('exist');
        });

        cy.get('h1').should('contain', 'Payment received');
        cy.checkA11yApp();
        cy.getCookie('pay').should('not.exist');

        cy.contains('a', 'Continue').click();

        cy.contains('.govuk-summary-list__row', 'Reference number').find('.govuk-summary-list__value')
            .invoke('text')
            .then((uid) => {
                cy.origin('http://localhost:9001', { args: { uid } }, ({ uid }) => {
                    cy.visit(`/?detail-type=payment-received&detail=${uid}`);
                    cy.contains('"amount":8200');
                    cy.contains('"paymentId":"hu20sqlact5260q2nanm0q8u93"');
                });
            });
    });

    it('can apply for a previous application fee reduction', () => {
        cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'Paying for your LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/are-you-applying-for-fee-discount-or-exemption')
        cy.checkA11yApp();

        cy.get('input[name="yes-no"]').check('yes', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/which-fee-type-are-you-applying-for')
        cy.checkA11yApp();

        cy.get('input[name="fee-type"]').check('RepeatApplicationFee', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/previous-application-number')
        cy.checkA11yApp();

        cy.get('#f-previous-application-number').invoke('val', '7ABC');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contains', '/how-much-did-you-previously-pay-for-your-lpa')
        cy.checkA11yApp();

        cy.get('input[name="selected"]').check('Half', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/evidence-required')
        cy.checkA11yApp();

        cy.get('h1').should('contain', 'New evidence required to pay a half fee');
        cy.contains('a', 'Continue').click();

        cy.url().should('contains', '/how-would-you-like-to-send-evidence')
        cy.checkA11yApp();

        cy.get('input[name="selected"]').check('post', { force: true });
        cy.contains('button', 'Continue').click();

        cy.url().should('contains', '/send-us-your-evidence-by-post')
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/payment-successful');
        cy.checkA11yApp();
        cy.contains('a', 'Continue').click()

        cy.contains('a', 'Return to task list').click()
        cy.url().should('contain', '/task-list');
        cy.contains('li', "Pay for the LPA").should('contain', 'Pending').click();

        cy.url().should('contain', '/pending-payment');
        cy.checkA11yApp();
        cy.contains('Repeat application, half fee (remission)');
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/which-fee-type-are-you-applying-for&progress=checkAndSendToYourCertificateProvider');

        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select which fee type you are applying for');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select which fee type you are applying for');
    });
});
