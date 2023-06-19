describe('Provide the certificate', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/provide-certificate&lpa.complete=1&certificateProviderProvided=1&loginAs=certificate-provider');
    });

    it('can provide the certificate', () => {
        cy.checkA11yApp();

        cy.get('#f-agree-to-statement').check()

        cy.contains('button', 'Submit signature').click();
        cy.url().should('contain', '/certificate-provided');
    });

    it("errors when not selected", () => {
        cy.contains('button', 'Submit signature').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the box to sign as the certificate provider');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Select the box to sign as the certificate provider');
    })
});
