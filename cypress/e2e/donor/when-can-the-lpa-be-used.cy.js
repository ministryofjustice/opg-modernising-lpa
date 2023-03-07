describe('When can the LPA be used', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used&withDonorDetails=1&withAttorney=1');
    });

    it('can be submitted', () => {
        cy.get('#f-when').check('when-registered');

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/restrictions');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select when your attorneys can use your LPA');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select when your attorneys can use your LPA');
    });
});
