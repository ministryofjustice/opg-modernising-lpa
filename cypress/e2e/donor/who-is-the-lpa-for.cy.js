describe('Who is the lpa for', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/who-is-the-lpa-for');
    });

    it('can be submitted', () => {
        cy.get('#f-who-for').check('me');

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/lpa-type');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select who the LPA is for');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select who the LPA is for');
    });
});
