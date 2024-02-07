describe('Dashboard', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?redirect=/supporter-dashboard&organisation=1');
    });

    it('can create a new LPA', () => {
        cy.checkA11yApp();
        cy.contains('button', 'Make a new LPA').click();

        cy.url().should('contain', '/your-details');
    });
});
