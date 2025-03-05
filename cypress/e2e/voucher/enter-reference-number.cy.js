const { TestEmail, randomShareCode } = require("../../support/e2e");

describe('Enter access code', () => {
    let shareCode = ''
    beforeEach(() => {
        shareCode = randomShareCode()

        cy.visit(`/fixtures/voucher?redirect=&withShareCode=${shareCode}`);

        cy.contains('a', 'Start').click()
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Random value').click();
            cy.contains('button', 'Continue').click();
        });
        cy.url().should('contain', '/voucher-enter-reference-number')
    });

    it('can enter a valid access code', { pageLoadTimeout: 6000 }, () => {
        cy.checkA11yApp();

        cy.get('#f-reference-number').invoke('val', shareCode);
        cy.contains('Save and continue').click();

        cy.url().should('contain', '/task-list')
    });

    it('errors when empty number', () => {
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter your access code');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'Enter your access code');
    });

    it('errors when incorrect code', () => {
        cy.get('#f-reference-number').invoke('val', 'i-am-very-wrong');
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The access code you entered is incorrect, please check it and try again');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The access code you entered is incorrect, please check it and try again');
    });

    it('errors when incorrect code length', () => {
        cy.get('#f-reference-number').invoke('val', 'short');
        cy.contains('Save and continue').click();

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('The access code you enter must be 12 characters');
        });

        cy.contains('[for=f-reference-number] ~ .govuk-error-message', 'The access code you enter must be 12 characters');
    });

});
