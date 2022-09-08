describe('About payment', () => {
    it('has a title', () => {
        cy.visit('/testing-start?redirect=/about-payment');
        cy.injectAxe();
        cy.get('h1').should('contain', 'About payment');
    })

    it('has a continue button', () => {
        cy.visit('/testing-start?redirect=/about-payment');
        cy.injectAxe();
        cy.contains('a', 'Continue to payment');
    })
})
