describe('Provide the certificate', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/provide-certificate&completeLpa=1&asCertificateProvider=1');
    });

    it('can provide the certificate', () => {
        cy.checkA11yApp();

        cy.get('#f-agree-to-statement').check()

        cy.contains('button', 'Confirm').click();
        cy.url().should('contain', '/certificate-provided');
    });

    it("errors when not selected", () => {
        cy.contains('button', 'Confirm').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select agree to statement');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Select agree to statement');
    })
});
