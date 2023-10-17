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

        cy.get('.govuk-notification-banner--success').within(() => {
            cy.contains('2 files successfully uploaded');
        });

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('dummy.pdf');
            cy.contains('dummy.png');
        });

        cy.contains('button', 'Continue to payment').click()

        cy.url().should('contain', '/payment-confirmation');
    })

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

        cy.get('.govuk-notification-banner--success').within(() => {
            cy.contains('1 file successfully uploaded');
        });

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('dummy.pdf');
        });

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/what-happens-after-no-fee');
        cy.checkA11yApp();
    })

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

        cy.get('.govuk-notification-banner--success').within(() => {
            cy.contains('1 file successfully uploaded');
        });

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('dummy.pdf');
        });

        cy.contains('button', 'Continue').click()

        // TODO: update once designs for page are finalised
        cy.url().should('contain', '/what-happens-after-no-fee');
        cy.checkA11yApp();
    })

    it('can only delete evidence that has not been sent to OPG', () => {
        cy.visit('/fixtures?redirect=/upload-evidence&progress=payForTheLpa&feeType=half-fee');
        cy.checkA11yApp();

        cy.get('input[type="file"]').attachFile(['dummy.pdf']);

        cy.contains('button', 'Upload files').click()

        cy.url().should('contain', '/upload-evidence');

        cy.get('.govuk-summary-list').within(() => {
            cy.contains('supporting-evidence.png').parent().should('not.contain', 'Delete');
            cy.contains('dummy.pdf').parent().contains('button', 'Delete').click();
        });

        cy.url().should('contain', '/upload-evidence');
        cy.checkA11yApp();

        cy.get('.moj-banner').within(() => {
            cy.contains('You have deleted file dummy.pdf');
        });
    })
});
