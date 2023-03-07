describe('Select your identity options', () => {
    beforeEach(() => {
        cy.visit('/testing-start?paymentComplete=1&redirect=/select-your-identity-options');
    });

    it('can select on first page', () => {
        cy.checkA11yApp();

        cy.contains('label', 'Your GOV.UK One Login Identity').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/your-chosen-identity-options');
        cy.checkA11yApp();

        cy.contains('Your GOV.UK One Login Identity');
        cy.contains('button', 'Continue');
    });

    it('can select on second page', () => {
        cy.checkA11yApp();

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/select-your-identity-options-1');
        cy.checkA11yApp();

        cy.contains('label', 'Your passport').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/your-chosen-identity-options');
        cy.checkA11yApp();

        cy.contains('passport');
        cy.contains('button', 'Continue');
    });

    it('can select on third page', () => {
        cy.checkA11yApp();

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/select-your-identity-options-1');
        cy.checkA11yApp();

        cy.contains('label', 'I do not have any of these types of identity').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/select-your-identity-options-2');
        cy.checkA11yApp();

        cy.contains('label', 'A bank account').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/your-chosen-identity-options');
        cy.checkA11yApp();

        cy.contains('your bank account');
        cy.contains('button', 'Continue');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select from the listed options');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select from the listed options');
    });
});
