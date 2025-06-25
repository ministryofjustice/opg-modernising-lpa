import { randomShareCode } from '../../support/e2e';

describe('Enter access code', () => {
    let accessCode = randomShareCode();

    beforeEach(() => {
        cy.visit(`/fixtures/supporter?redirect=/enter-access-code&organisation=1&accessCode=${accessCode}`);
    });

    it('links the LPA', () => {
        cy.checkA11yApp();
        cy.get('#f-reference-number').invoke('val', accessCode);
        cy.contains('button', 'Continue').click();

        cy.contains('M-FAKE-');
        cy.contains('a', 'Go to task list').click();
        cy.contains('LPA task list');
    });
});
