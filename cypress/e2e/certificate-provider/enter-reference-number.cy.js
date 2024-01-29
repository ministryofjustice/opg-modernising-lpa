const { TestEmail } = require("../../support/e2e");

describe('Enter reference number', () => {
    beforeEach(() => {
        cy.visit(`/fixtures/certificate-provider?redirect=/certificate-provider-start&use-test-code=1&email=${TestEmail}`);

        cy.contains('a', 'Start').click()
        cy.contains('button', 'Sign in').click();
        cy.url().should('contain', '/certificate-provider-enter-reference-number')
    });

    it('can enter a valid reference number', { pageLoadTimeout: 6000 }, () => {
        cy.checkA11yApp();

        cy.get('#f-reference-number').type('1234-5678');
        cy.contains('Save and continue').click();

        cy.url().should('contain', '/certificate-provider-who-is-eligible')
    });

    it('errors when empty number', () => {
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your 8 character reference number');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'Enter your 8 character reference number');
    });

    it('errors when incorrect code', () => {
        cy.get('#f-reference-number').type('not-right');
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The reference number you entered is incorrect, please check it and try again');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The reference number you entered is incorrect, please check it and try again');
    });

    it('errors when incorrect code length', () => {
        cy.get('#f-reference-number').type('short');
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The reference number you enter must be 8 characters');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The reference number you enter must be 8 characters');
    });

});
