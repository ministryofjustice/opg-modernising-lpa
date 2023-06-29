describe('data loss warning accessibility', () => {
    it('locks focus to data loss warning dialog', () => {
        cy.visit('/testing-start?redirect=/choose-attorneys&lpa.yourDetails=1&cookiesAccepted=1');

        cy.get('#f-first-names').type('John');
        cy.contains('a', 'Return to task list').click()

        cy.get('#dialog').should('be.visible')
        cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none')

        cy.focused().should('have.attr', 'id', 'back-to-page-btn')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'return-to-task-list-dialog-btn')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'dialog-title')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'dialog-description')

        cy.realPress("Tab")
        cy.focused().should('have.attr', 'id', 'back-to-page-btn')

        cy.contains('button', 'Back to page').click()

        cy.get('#dialog').should('not.be.visible')
        cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none')

        cy.focused().should('have.attr', 'id', 'return-to-tasklist-btn')

        cy.contains('a', 'Return to task list').click()

        cy.get('#dialog').should('be.visible')
        cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none')

        cy.focused().should('have.attr', 'id', 'back-to-page-btn')

        cy.realPress(["Shift", "Tab"])
        cy.focused().should('have.attr', 'id', 'dialog-description')

        cy.realPress(["Shift", "Tab"])
        cy.focused().should('have.attr', 'id', 'dialog-title')

        cy.realPress(["Shift", "Tab"])
        cy.focused().should('have.attr', 'id', 'return-to-task-list-dialog-btn')

        cy.realPress(["Shift", "Tab"])
        cy.focused().should('have.attr', 'id', 'back-to-page-btn')

        cy.realType('{esc}')

        cy.get('#dialog').should('not.be.visible')
        cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none')

        cy.focused().should('have.attr', 'id', 'return-to-tasklist-btn')
    })
})
