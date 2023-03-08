describe('Start', () => {
    beforeEach(() => {
        cy.visit('/');
    });

    afterEach(() => {
        cy.checkA11yApp();
    });

    it('has a title', () => {
        cy.get('h1').should('contain', 'Make a lasting power of attorney');
    });

    it('has a start button', () => {
        cy.contains('a', 'Start');
    });
});
