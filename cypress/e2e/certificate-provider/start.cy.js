const { oneLoginUrl, isLocal } = require("../../support/e2e");

describe('Start', () => {
    beforeEach(() => {
        cy.visit('/certificate-provider-start');
    });

    it('can be completed', () => {
        cy.contains('a', 'Start').click();

        if (isLocal()) {
            cy.origin(oneLoginUrl(), () => {
                cy.url().should('contain', '/authorize')
            });
        } else {
            cy.origin('https://signin.integration.account.gov.uk', () => {
                cy.url().should('contain', '/')
            })
        }
    });
});
