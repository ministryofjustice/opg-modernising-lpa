describe('Start', () => {
    beforeEach(() => {
        cy.visit('/');
    });

    afterEach(() => {
        cy.checkA11yApp();
    });

    it('has a title', () => {
        cy.get('h1').should('contain', 'Make and register a lasting power of attorney (LPA)');
    });

    it('has a start button', () => {
        cy.contains('a', 'Start');
    });
});
