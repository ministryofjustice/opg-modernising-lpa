describe('Previous application number', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/previous-application-number');
    });

    it('can be submitted', () => {
        cy.checkA11yApp();

        cy.get('#f-previous-application-number').type('MABC');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/what-happens-after-no-fee');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter previous reference number');
        });

        cy.contains('.govuk-error-message', 'Enter previous reference number');
    });

    it('errors when not correct format', () => {
        cy.get('#f-previous-application-number').type('ABC');
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Previous reference number must begin with 7 or M');
        });

        cy.contains('.govuk-error-message', 'Previous reference number must begin with 7 or M');
    });
});
