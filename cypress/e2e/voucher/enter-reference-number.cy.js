const { TestEmail, randomShareCode } = require("../../support/e2e");

describe('Enter reference number', () => {
    let shareCode = ''
    beforeEach(() => {
        shareCode = randomShareCode()

        cy.visit(`/fixtures/voucher?redirect=/voucher-start&withShareCode=${shareCode}&email=${TestEmail}`);

        cy.contains('a', 'Start').click()
        cy.contains('label', 'Random value').click();
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/voucher-enter-reference-number')
    });

    it('can enter a valid reference number', { pageLoadTimeout: 6000 }, () => {
        cy.checkA11yApp();

        cy.get('#f-reference-number').type(shareCode);
        cy.contains('Save and continue').click();

        cy.url().should('contain', '/task-list')
    });

    it('errors when empty number', () => {
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your 12 character reference number');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'Enter your 12 character reference number');
    });

    it('errors when incorrect code', () => {
        cy.get('#f-reference-number').type('i-am-very-wrong');
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
            cy.contains('The reference number you enter must be 12 characters');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The reference number you enter must be 12 characters');
    });

});
