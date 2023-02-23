describe('Enter reference code', () => {
    beforeEach(() => {
        cy.visit('/testing-start?completeLpa=1&startCpFlowWithId=1');
    });

    it('can enter a valid reference code', () => {
        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-reference-code').type('abcdef123456');
        cy.contains('Continue').click();

        cy.url().should('contain', '/certificate-provider-login-callback');
    });

    it('errors when empty code', () => {
        cy.contains('a', 'Start').click()

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter reference code');
        });

        cy.contains('[for=f-reference-code] + .govuk-error-message', 'Enter reference code');
    });
});
