describe('How should replacement attorneys make decisions', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/how-should-replacement-attorneys-make-decisions');
    });

    it('can choose how replacement attorneys act', () => {
        cy.contains('h1', 'How should your replacement attorneys make decisions?');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="decision-type"]').check('jointly');

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/task-list');
    });

    it('can choose how replacement attorneys act - Jointly for some decisions, and jointly and severally for other decisions', () => {
        cy.get('input[name="decision-type"]').check('jointly-for-some-severally-for-others');
        cy.get('#f-mixed-details').type('some details on attorneys');

        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/task-list');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select how the replacement attorneys should make decisions');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select how the replacement attorneys should make decisions');
    });

    it('errors when details empty', () => {
        cy.get('input[name="decision-type"]').check('jointly-for-some-severally-for-others');
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter details');
        });

        cy.contains('[for=f-mixed-details] + .govuk-error-message', 'Enter details');
    });
});
