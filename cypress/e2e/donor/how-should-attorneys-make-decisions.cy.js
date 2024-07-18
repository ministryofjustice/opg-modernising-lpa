describe('How should attorneys make decisions', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/how-should-attorneys-make-decisions');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });
    });

    it('can choose how attorneys act - Jointly and severally', () => {
        cy.contains('h1', 'How should your attorneys make decisions?');

        cy.get('input[name="decision-type"]').check('jointly-and-severally', { force: true });

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/task-list');
    });

    it('can choose how attorneys act - Jointly', () => {
        cy.contains('h1', 'How should your attorneys make decisions?');

        cy.get('input[name="decision-type"]').check('jointly', { force: true });

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/because-you-have-chosen-jointly');
        cy.contains('a', 'Return to task list').click();

        cy.url().should('contain', '/task-list');
    });

    it('can choose how attorneys act - Jointly for some decisions, and jointly and severally for other decisions', () => {
        cy.contains('h1', 'How should your attorneys make decisions?');

        cy.get('input[name="decision-type"]').check('jointly-for-some-severally-for-others', { force: true });
        cy.get('#f-mixed-details').type('some details on attorneys');

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/because-you-have-chosen-jointly-for-some-severally-for-others');
        cy.contains('a', 'Return to task list').click();

        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how the attorneys should make decisions');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how the attorneys should make decisions');
    });

    it('errors when details empty', () => {
        cy.get('input[name="decision-type"]').check('jointly-for-some-severally-for-others', { force: true });
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter details');
        });

        cy.contains('[for=f-mixed-details] + .govuk-error-message', 'Enter details');
    });
});
