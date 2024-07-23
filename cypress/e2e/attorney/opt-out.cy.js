const { TestEmail, randomShareCode } = require("../../support/e2e");

describe('Opting out', () => {
    let shareCode = ''

    it('stops me being attorney', () => {
        shareCode = randomShareCode()
        cy.visit(`/fixtures/attorney?redirect=&withShareCode=${shareCode}&email=${TestEmail}`);

        cy.visit('/attorney-enter-reference-number-opt-out');
        cy.checkA11yApp();
        cy.get('#f-reference-number').type(shareCode);
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-you-do-not-want-to-be-an-attorney')
        cy.checkA11yApp();
        cy.contains('M-FAKE-');

        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/you-have-decided-not-to-be-an-attorney');
    });
});
