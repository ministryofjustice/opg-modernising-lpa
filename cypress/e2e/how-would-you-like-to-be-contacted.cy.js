describe('How would you like to be contacted', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/how-would-you-like-to-be-contacted');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-contact').check('email');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/start');
    });
});
