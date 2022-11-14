describe('How should replacement attorneys make decisions', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/how-should-replacement-attorneys-make-decisions?cookiesAccepted=1');
        cy.injectAxe();
    });

    it('can choose how replacement attorneys act', () => {
        cy.contains('h1', 'How should the replacement attorneys make decisions?');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="decision-type"]').check('jointly');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

    });

    it('can choose how replacement attorneys act - Jointly for some decisions, and jointly and severally for other decisions', () => {
        cy.contains('h1', 'How should the replacement attorneys make decisions?');

        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="decision-type"]').check('mixed');
        cy.get('#mixed-details').type('some details on attorneys');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });
    });
});
