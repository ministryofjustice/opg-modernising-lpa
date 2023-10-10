describe('Payment', () => {
    describe('Pay for LPA', () => {
        it('adds a secure cookie before redirecting user to GOV UK Pay', () => {
            cy.clearCookie('pay');
            cy.getCookie('pay').should('not.exist')

            cy.visit('/fixtures?redirect=/about-payment&progress=checkAndSendToYourCertificateProvider');
            cy.checkA11yApp();

            cy.get('h1').should('contain', 'About payment');
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
    });
});
