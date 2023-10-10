describe('Previous application number', () => {
    it('can be submitted', () => {
        cy.visit('/fixtures?redirect=/previous-application-number');
        cy.checkA11yApp();

        cy.get('#f-previous-application-number').type('ABC');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.visit('/fixtures?redirect=/previous-application-number');

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter previousApplicationNumber');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Enter previousApplicationNumber');
    });
});
