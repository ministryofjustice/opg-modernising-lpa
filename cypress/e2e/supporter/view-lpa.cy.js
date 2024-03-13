describe('View LPA', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/dashboard&lpa=1');
        cy.checkA11yApp();
    });

    it('can continue making an LPA', () => {
        cy.contains('a', 'M-FAKE').click()

        cy.url().should('contain', '/view-lpa');
        cy.checkA11yApp();

        cy.contains('h1', 'Property and affairs LPA')
        cy.contains('div', 'M-FAKE')

        cy.contains('a', 'Donor access')
        cy.contains('a', 'View LPA summary')
        cy.contains('a', 'Go to task list').click()

        cy.url().should('contain', '/task-list');
        cy.checkA11yApp();

        cy.contains('Provide your details').click()

        cy.url().should('contain', '/your-details');
        cy.checkA11yApp();

        cy.get('#f-first-names').type('2');
        cy.contains('button', 'Continue').click();
        cy.contains('a', 'Dashboard').click();
        cy.contains('Sam2 Smith');
    });
})
