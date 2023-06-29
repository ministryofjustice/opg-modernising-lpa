describe('data loss warning accessibility', () => {
    it('locks focus to data loss warning dialog', () => {
        cy.visit('/testing-start?redirect=/choose-attorneys&lpa.yourDetails=1&cookiesAccepted=1');

        cy.get('#f-first-names').type('John');
        cy.contains('a', 'Return to task list').click()

        cy.contains('#dialog').should('be.visible')
        cy.contains('#dialog-overlay').should('be.visible')

        cy.focused().should('have.attr', 'id', 'dialog-focus')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'back-to-page')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'return-to-task-list-dialog')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'back-to-page')

        cy.contains('button', 'Back to page').click()

        cy.contains('#dialog').should('not.be.visible')
        cy.contains('#dialog-overlay').should('not.be.visible')

        cy.focused().should('have.text', 'Return to task list')

        cy.contains('a', 'Return to task list').click()

        cy.contains('#dialog').should('be.visible')
        cy.contains('#dialog-overlay').should('be.visible')

        cy.focused().should('have.attr', 'id', 'dialog-focus')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'back-to-page')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'return-to-task-list-dialog')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'back-to-page')

        cy.realType('{esc}')

        cy.contains('#dialog').should('not.be.visible')
        cy.contains('#dialog-overlay').should('not.be.visible')

        cy.focused().should('have.text', 'Return to task list')
    })
})
