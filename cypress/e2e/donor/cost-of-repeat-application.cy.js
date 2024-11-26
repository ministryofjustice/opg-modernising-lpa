describe('Cost of repeat application', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/cost-of-repeat-application');
    });

    it('can be submitted', () => {
        cy.checkA11yApp();

        cy.contains('label', 'no fee').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/what-happens-next-repeat-application-no-fee');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select which fee you are eligible to pay');
        });

        cy.contains('.govuk-error-message', 'Select which fee you are eligible to pay');
    });
});
