describe('Who is the lpa for', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/lpa-type');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-lpa-type').check('pfa');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/who-is-the-lpa-for');
    });
});
