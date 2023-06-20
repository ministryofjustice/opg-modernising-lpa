describe('Check your name', () => {
    it('can confirm name matches', () => {
        cy.visit('/testing-start?redirect=/check-your-name&lpa.complete=1&certificateProviderProvided=1&loginAs=certificate-provider');

        // see https://github.com/alphagov/govuk-frontend/issues/979
        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('input[name="is-name-correct"]').check('yes');
        cy.contains('Continue').click();

        cy.url().should('contain', '/enter-date-of-birth');
    });

    it('can provide an updated name', () => {
        cy.visit('/testing-start?redirect=/check-your-name&lpa.complete=1&certificateProviderProvided=1&loginAs=certificate-provider');

        cy.get('input[name="is-name-correct"]').check('no');
        cy.get('#f-corrected-name').type('New Name');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.contains('Continue').click();

        cy.url().should('contain', '/enter-date-of-birth');
    });

    it('errors when not selected', () => {
        cy.visit('/testing-start?redirect=/check-your-name&lpa.complete=1&certificateProviderProvided=1&loginAs=certificate-provider');

        cy.contains('Continue').click();

        cy.url().should('contain', '/check-your-name');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes if the name is correct');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Select yes if the name is correct');
    });

    it('errors when name not correct but no name provided', () => {
        cy.visit('/testing-start?redirect=/check-your-name&lpa.complete=1&certificateProviderProvided=1&loginAs=certificate-provider');

        cy.get('input[name="is-name-correct"]').check('no');
        cy.contains('Continue').click();

        cy.url().should('contain', '/check-your-name');

        cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your full name');
        });

        cy.contains('[for=f-corrected-name] ~ .govuk-error-message', 'Enter your full name');
    });
});
