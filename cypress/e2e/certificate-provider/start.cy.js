describe('Start', () => {
    beforeEach(() => {
        cy.visit('/testing-start?startCpFlowDonorHasPaid=1');
    });

    it('can be completed', () => {
        cy.contains('a', 'Start').click();

        if (Cypress.config().baseUrl.includes('localhost')) {
            cy.url().should('contain', '/certificate-provider-enter-reference-number')
        } else {
            cy.origin('https://signin.integration.account.gov.uk', () => {
                cy.url().should('contain', '/')
            })
        }
    });
});
