describe('View LPA', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/dashboard&lpa=1');
        cy.checkA11yApp();
    });

    it('shows LPA reference number and progress tracker', () => {
        cy.contains('a', 'M-FAKE').click()

        cy.url().should('contain', '/view-lpa');
        cy.checkA11yApp();

        cy.contains('h1', 'Property and affairs LPA')
        cy.contains('p', 'M-FAKE')

        cy.contains('h2', 'LPA progress')
        cy.contains('li', 'Sam Smith has paid');
    });
})
