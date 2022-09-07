describe('Restrictions', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/restrictions');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-restrictions').type('this that');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });
});
