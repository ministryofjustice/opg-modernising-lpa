describe('Payment', () => {
    describe('Pay for LPA', () => {
        it('adds a secure cookie before redirecting user to GOV UK Pay', () => {
            cy.getCookie('pay').should('not.exist')

            cy.visit('/testing-start?redirect=/about-payment');
            cy.injectAxe();

            cy.get('h1').should('contain', 'About payment');

            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.intercept('**/v1/payments', (req) => {
                cy.getCookie('pay').should('exist')
            })

            cy.contains('button', 'Continue to payment').click()
        })

        it('removes existing secure cookie on payment confirmation page', () => {
            cy.visit('/testing-start?redirect=/payment-confirmation&paymentComplete=1');

            cy.injectAxe();

            cy.get('h1').should('contain', 'Payment received');
            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.contains('a', 'Continue').click()

            cy.url().should('contains', '/select-your-identity')

            cy.getCookie('pay').should('not.exist')
        })
    })
})
