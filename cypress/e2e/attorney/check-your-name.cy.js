describe('Check your name', () => {
    it('can confirm name matches', () => {
        cy.visit('/testing-start?redirect=/attorney-check-your-name&lpa.complete=1&attorneyProvided=1&loginAs=attorney');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="is-name-correct"]').check('yes');
        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-read-the-lpa');
    });

    it('can provide an updated name', () => {
        cy.visit('/testing-start?redirect=/attorney-check-your-name&lpa.complete=1&attorneyProvided=1&loginAs=attorney');

        cy.get('input[name="is-name-correct"]').check('no');
        cy.get('#f-corrected-name').type('New Name');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-read-the-lpa');
    });

    it('errors when not selected', () => {
        cy.visit('/testing-start?redirect=/attorney-check-your-name&lpa.complete=1&attorneyProvided=1&loginAs=attorney');

        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-check-your-name');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Confirm if the name the donor provided for you is correct');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Confirm if the name the donor provided for you is correct');
    });

    it('errors when name not correct but no name provided', () => {
        cy.visit('/testing-start?redirect=/attorney-check-your-name&lpa.complete=1&attorneyProvided=1&loginAs=attorney');

        cy.get('input[name="is-name-correct"]').check('no');
        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-check-your-name');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your full name');
        });

        cy.contains('[for=f-corrected-name] ~ .govuk-error-message', 'Enter your full name');
    });
});
