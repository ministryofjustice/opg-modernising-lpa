describe('Delete LPA', () => {
    it('can be deleted', () => {
        cy.visit('/fixtures?redirect=&progress=provideYourDetails');

        cy.contains('Sam Smith');
        cy.contains('a', 'Delete LPA').click();

        cy.checkA11yApp();
        cy.contains('button', 'Delete this LPA').click();

        cy.checkA11yApp();
        cy.contains('has been deleted');
        cy.contains('a', 'Return to dashboard').click();

        cy.contains('Sam Smith').should('not.exist');
    });
});
