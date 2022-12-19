describe('Read your LPA', () => {
    it('continues to sign the LPA', () => {
        cy.visit('/testing-start?redirect=/read-your-lpa');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('h1', 'Read and sign your LPA')
        cy.contains('a', 'Continue to signing page').click()

        cy.url().should('contain', '/sign-your-lpa');
    });
});
