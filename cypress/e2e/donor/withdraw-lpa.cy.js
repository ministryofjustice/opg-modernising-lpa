describe('Withdraw LPA', () => {
    it('can be withdrawn', () => {
        cy.visit('/fixtures?redirect=&progress=submitted');

        cy.contains('Sam Smith');
        cy.contains('a', 'Withdraw LPA').click();

        cy.checkA11yApp();
        cy.contains('button', 'Withdraw this LPA').click();

        cy.checkA11yApp();
        cy.contains('You have withdrawn');
        cy.contains('a', 'Return to dashboard').click();

        cy.contains('.app-dashboard-card', 'Sam Smith').contains('.app-tag', 'Withdrawn');
    });
});
