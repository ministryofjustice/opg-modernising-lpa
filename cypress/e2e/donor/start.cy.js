describe('Start', () => {
    beforeEach(() => {
        cy.visit('/');
        cy.injectAxe();
    });

    afterEach(() => {
        cy.checkA11y(null, {
            rules: { region: { enabled: false } },
        });
    });

    it('has a title', () => {
        cy.get('h1').should('contain', 'Make a lasting power of attorney');
    });

    it('has a start button', () => {
        cy.contains('a', 'Start');
    });
});
