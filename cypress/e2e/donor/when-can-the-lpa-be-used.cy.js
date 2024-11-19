describe('When can the LPA be used', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/when-can-the-lpa-be-used&progress=chooseYourAttorneys');
    });

    it('can be submitted', () => {
        cy.get('#f-selected').check('when-has-capacity', { force: true });

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select when your attorneys can use your LPA');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select when your attorneys can use your LPA');
    });
});
