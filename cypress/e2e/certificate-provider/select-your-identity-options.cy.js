describe('Select your identity options', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-select-your-identity-options&completeLpa=1&asCertificateProvider=1');
    });

    it('can select on first page', () => {
        cy.checkA11yApp();

        cy.contains('label', 'Your GOV.UK One Login Identity').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/certificate-provider-your-chosen-identity-options');
        cy.checkA11yApp();

        cy.contains('Your GOV.UK One Login Identity');
        cy.contains('button', 'Continue');
    });

    it('can select on second page', () => {
        cy.checkA11yApp();

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/certificate-provider-select-identity-document');
        cy.checkA11yApp();

        cy.contains('label', 'Your passport').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/certificate-provider-your-chosen-identity-options');
        cy.checkA11yApp();

        cy.contains('passport');
        cy.contains('button', 'Continue');
    });

    it('can select on third page', () => {
        cy.checkA11yApp();

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/certificate-provider-select-identity-document');
        cy.checkA11yApp();

        cy.contains('label', 'I do not have any of these identity documents').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/certificate-provider-select-identity-document-2');
        cy.checkA11yApp();

        cy.contains('label', 'A bank account').click();
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/certificate-provider-your-chosen-identity-options');
        cy.checkA11yApp();

        cy.contains('your bank account');
        cy.contains('button', 'Continue');
    });

    it('errors when unselected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select from the listed options');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select from the listed options');
    });
});
