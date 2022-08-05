describe('GDS and MOJ components are available', () => {
    it('displays a GDS summary element', () => {
        cy.visit('http://localhost:5050/home')

        cy.get('summary').first()
            .should('contain.text', 'Help with nationality')
    })

    it('displays a MOJ password reveal element', () => {
        cy.visit('http://localhost:5050/home')

        cy.get('[data-module=moj-password-reveal]').first()
            .should('have.value', '1234ABC!')
            .should('have.attr', 'type', 'password')

        cy.get('button').contains('Show')
            .click()

        cy.get('[data-module=moj-password-reveal]').first()
            .should('have.attr', 'type', 'text')
    })
})
