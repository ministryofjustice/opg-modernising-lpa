describe('Payment', () => {
    describe('Pay for LPA', () => {
        it('adds a secure cookie before redirecting user to GOV UK Pay', () => {
            cy.clearCookie('pay');
            cy.getCookie('pay').should('not.exist')

            cy.visit('/testing-start?redirect=/task-list&lpa.certificateProvider=1');
            cy.visitLpa('/about-payment');

            cy.get('h1').should('contain', 'About payment');

            cy.checkA11yApp();

            cy.intercept('**/v1/payments', (req) => {
                cy.getCookie('pay').should('exist')
            })

            cy.contains('button', 'Continue to payment').click()
        })

        it('removes existing secure cookie on payment confirmation page', () => {
            cy.visit('/testing-start?redirect=/task-list&lpa.certificateProvider=1&lpa.paid=1');
            cy.getCookie('pay').should('exist')

            cy.visitLpa('/payment-confirmation');

            cy.get('h1').should('contain', 'Payment received');
            cy.checkA11yApp();

            cy.getCookie('pay').should('not.exist')

            cy.contains('a', 'Continue').click()

            cy.url().should('contains', '/task-list')
        })
    })
})
