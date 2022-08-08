describe('GDS and MOJ components are available', () => {
    beforeEach(() => {
        cy.visit('/home')
        cy.injectAxe()
    })

    it('displays a GDS summary element', () => {
        cy.checkA11y()

        cy.get('summary').first()
            .should('contain.text', 'Help with nationality')
    })

    it('displays a MOJ password reveal element', () => {
        cy.checkA11y()

        cy.get('[data-module=moj-password-reveal]').first()
            .should('have.value', '1234ABC!')
            .should('have.attr', 'type', 'password')

        cy.get('button').contains('Show')
            .click()

        cy.get('[data-module=moj-password-reveal]').first()
            .should('have.attr', 'type', 'text')

        cy.checkA11y()
    })
})
