describe('Provide the certificate', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/provide-certificate&progress=confirmYourIdentity');
    });

    it('can provide the certificate', () => {
        cy.checkA11yApp();

        cy.get('#f-agree-to-statement').check({ force: true })

        cy.contains('button', 'Submit signature').click();
        cy.url().should('contain', '/certificate-provided');
    });

    it('can choose not to provide the certificate', () => {
        cy.checkA11yApp();

        cy.contains('I cannot provide the certificate').click();

        cy.url().should('contain', '/confirm-you-do-not-want-to-be-a-certificate-provider')
        cy.checkA11yApp();

        cy.contains('Property and affairs')

        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/you-have-decided-not-to-be-a-certificate-provider')
        cy.checkA11yApp();

        cy.contains('You have confirmed that you do not want to be Sam Smith’s certificate provider')
        cy.contains('We have let Sam know about your decision.')
    });

    it("errors when not selected", () => {
        cy.contains('button', 'Submit signature').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the box to sign as the certificate provider');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Select the box to sign as the certificate provider');
    });

    it('errors when when the wrong language is used', () => {
        cy.contains('a', 'Cymraeg').click();
        cy.get('#f-agree-to-statement').check({ force: true })

        cy.contains('button', 'Cyflwyno llofnod').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('yn Saesneg');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'yn Saesneg');
    });
});
