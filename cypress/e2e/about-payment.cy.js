describe('About payment', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/about-payment');
        cy.injectAxe();
    });

    it('has a title and continue button', () => {
        cy.get('h1').should('contain', 'About payment');
        cy.contains('a', 'Continue to payment');
        cy.checkA11y(null, { rules: { region: { enabled: false } } });
    })
})
