describe('Withdraw LPA', () => {
    it('can be withdrawn', () => {
        cy.visit('/fixtures?redirect=&progress=statutoryWaitingPeriod');

        cy.contains('Sam Smith');
        cy.contains('a', 'Revoke LPA').click();

        cy.checkA11yApp();
        cy.contains('button', 'Revoke this LPA').click();

        cy.checkA11yApp();
        cy.contains('You have revoked');
        cy.contains('a', 'Return to dashboard').click();

        cy.contains('.app-dashboard-card', 'Sam Smith').contains('.app-tag', 'Revoked');
    });
});
