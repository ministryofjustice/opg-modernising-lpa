describe('Restrictions', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/restrictions');
    });

    it('can be submitted', () => {
        cy.get('#f-restrictions').type('this that');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/task-list');
    });
});
