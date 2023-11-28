const { TestEmail } = require("../../support/e2e");

describe('Enter reference number', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/certificate-provider-start&use-test-code=1&email=' + TestEmail);
    });

    it('can enter a valid reference number', { pageLoadTimeout: 6000 }, () => {
        cy.contains('a', 'Start').click()

        cy.checkA11yApp();

        cy.get('#f-reference-number').type('abcdef123456');
        cy.contains('Save and continue').click();

        cy.url().should('contain', '/certificate-provider-who-is-eligible')
    });

    it('errors when empty number', () => {
        cy.contains('a', 'Start').click()

        cy.checkA11yApp();

        cy.contains('Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your 12 character reference number');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'Enter your 12 character reference number');
    });

    it('errors when incorrect code', () => {
        cy.contains('a', 'Start').click()

        cy.checkA11yApp();

        cy.get('#f-reference-number').type('notATestCode');
        cy.contains('Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The reference number you entered is incorrect, please check it and try again');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The reference number you entered is incorrect, please check it and try again');
    });

    it('errors when incorrect code length', () => {
        cy.contains('a', 'Start').click()

        cy.checkA11yApp();

        cy.get('#f-reference-number').type('tooShort');
        cy.contains('Save and continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The reference number you enter must be 12 characters');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The reference number you enter must be 12 characters');
    });
});
