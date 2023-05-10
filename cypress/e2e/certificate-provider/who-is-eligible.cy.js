describe('Who is eligible', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-who-is-eligible&withShareCodeSession=1');
    });

    it('can continue', () => {
        cy.checkA11yApp();

        cy.contains('Continue').click();

        if (Cypress.config().baseUrl.includes('localhost')) {
            cy.url().should('contain', '/certificate-provider-enter-date-of-birth')
        } else {
            cy.origin('https://signin.integration.account.gov.uk', () => {
                cy.url().should('contain', '/')
            })
        }
    });
});
