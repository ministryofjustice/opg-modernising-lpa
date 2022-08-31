describe('Who is the lpa for', () => {
    beforeEach(() => {
        cy.visit('/auth');
        cy.visit('/who-is-the-lpa-for');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-who-for').check('me');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/donor-details');
    });
});
