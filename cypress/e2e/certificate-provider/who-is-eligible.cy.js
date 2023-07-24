describe('Who is eligible', () => {
    beforeEach(() => {
        cy.visit('/testing-start?withShareCodeSession=1');
        cy.visit('/certificate-provider-who-is-eligible');
    });

    it('can continue', () => {
        cy.checkA11yApp();

        cy.contains('Continue').click();

        if (Cypress.config().baseUrl.includes('localhost')) {
            cy.url().should('contain', '/task-list')
        } else {
            cy.origin('https://signin.integration.account.gov.uk', () => {
                cy.url().should('contain', '/')
            })
        }
    });
});
