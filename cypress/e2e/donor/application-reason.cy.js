describe('Application reason', () => {
    it('can be submitted', () => {
        cy.visit('/fixtures?redirect=/application-reason&progress=provideYourDetails');
        cy.checkA11yApp();

        cy.contains('label', 'noneOfTheAbove').click();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/application-reason&progress=provideYourDetails');

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select theReasonForMakingTheApplication');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select theReasonForMakingTheApplication');
    });
});
