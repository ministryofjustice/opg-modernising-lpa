describe('Restrictions', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/restrictions&lpa.yourDetails=1&lpa.attorneys=1');
    });

    it('can be submitted', () => {
        cy.get('#f-restrictions').type('this that');

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click();
        cy.url().should('contain', '/task-list');
    });
});
