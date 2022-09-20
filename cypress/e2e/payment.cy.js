describe('Payment', () => {
    describe('Call to action', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/about-payment');
            cy.injectAxe();
        });

        it('has a title and continue button', () => {
            cy.get('h1').should('contain', 'About payment');
            cy.contains('button', 'Continue to payment');
            cy.checkA11y(null, { rules: { region: { enabled: false } } });
        })
    })

    describe('Pay for LPA', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/about-payment');
        });

        it('adds a secure cookie before redirecting user to GOV UK Pay', () => {
            cy.getCookie('pay').should('not.exist')
            cy.contains('button', 'Continue to payment').click()

            cy.getCookie('pay').should('exist')
            cy.url().should('eq', `${Cypress.config('baseUrl')}/payment-confirmation`)
        })
    })
})
