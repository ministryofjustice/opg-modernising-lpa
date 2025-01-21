describe('Start', () => {
    beforeEach(() => {
        cy.visit('/certificate-provider-start');
    });

    it('can be completed', () => {
        cy.contains('a', 'Start').click();

        if (Cypress.config().baseUrl.includes('localhost')) {
            cy.origin('http://localhost:7012', () => {
                cy.url().should('contain', '/authorize')
            });
        } else {
            cy.origin('https://signin.integration.account.gov.uk', () => {
                cy.url().should('contain', '/')
            })
        }
    });
});
