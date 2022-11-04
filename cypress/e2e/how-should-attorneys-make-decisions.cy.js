describe('How should attorneys make decisions', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/how-should-attorneys-make-decisions?cookiesAccepted=1');
        cy.injectAxe();
    });

    it('can choose how attorneys act', () => {
        cy.contains('h1', 'How should the attorneys make decisions?');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="decision-type"]').check('jointly');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/want-replacement-attorneys');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

    });

    it('can choose how attorneys act - Jointly for some decisions, and jointly and severally for other decisions', () => {
        cy.contains('h1', 'How should the attorneys make decisions?');

        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="decision-type"]').check('mixed');
        cy.get('#mixed-details').type('some details on attorneys');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/want-replacement-attorneys');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });
    });
});
