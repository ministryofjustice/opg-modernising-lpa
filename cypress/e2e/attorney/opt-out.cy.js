const { TestEmail, randomAccessCode } = require("../../support/e2e");

describe('Opting out', () => {
    it('stops me being attorney', () => {
        const accessCode = randomAccessCode();
        cy.visit(`/fixtures/attorney?redirect=&withAccessCode=${accessCode}&email=${TestEmail}`);

        cy.visit('/attorney-enter-access-code-opt-out');
        cy.checkA11yApp();
        cy.get('#f-donor-last-name').type('Smith');
        cy.get('#f-access-code').invoke('val', accessCode);
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-you-do-not-want-to-be-an-attorney');
        cy.checkA11yApp();
        cy.contains('M-FAKE-');

        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/you-have-decided-not-to-be-an-attorney');
    });
});
