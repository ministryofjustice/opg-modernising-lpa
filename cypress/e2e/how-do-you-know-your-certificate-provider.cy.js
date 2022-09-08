describe('How would you like to be contacted', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/how-do-you-know-your-certificate-provider');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-how').check('friend');

        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });
});
