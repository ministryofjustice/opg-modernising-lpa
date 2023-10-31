describe('LPA type', () => {
    it('can be submitted', () => {
        cy.visit('/fixtures?redirect=/lpa-type&progress=provideYourDetails');

        cy.get('#f-lpa-type').check('pfa');

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/lpa-type');

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the type of LPA to make');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select the type of LPA to make');
    });
});
