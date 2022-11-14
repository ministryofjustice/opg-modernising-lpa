describe('How should replacement attorneys step in', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/how-should-replacement-attorneys-step-in?cookiesAccepted=1');
        cy.injectAxe();
    });

    it('can choose how replacement attorneys step in', () => {
        cy.contains('h1', 'How should the replacement attorneys step in?');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="when-to-step-in"]').check('one-can-no-longer-act');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

    });

    it('can choose how replacement attorneys step in - some other way', () => {
        cy.contains('h1', 'How should the replacement attorneys step in?');

        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="when-to-step-in"]').check('other');
        cy.get('#other-details').type('some details on when to step in');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });
    });
});
