describe('Application reason', () => {
    it('can be submitted', () => {
        cy.visit('/testing-start?redirect=/application-reason&lpa.yourDetails=1');
        cy.checkA11yApp();

        cy.contains('label', 'additionalApplication').click();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.visit('/testing-start?redirect=/application-reason');

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select theReasonForMakingTheApplication');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select theReasonForMakingTheApplication');
    });
});
