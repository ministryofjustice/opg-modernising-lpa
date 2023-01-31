describe('When can the LPA be used', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used&withAttorney=1');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-when').check('when-registered');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/restrictions');
    });
    
    it('errors when unselected', () => {
        cy.contains('button', 'Continue').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select when the LPA can be used');
        });
        
        cy.contains('.govuk-fieldset .govuk-error-message', 'Select when the LPA can be used');
    });
});
