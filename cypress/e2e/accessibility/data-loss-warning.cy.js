describe('Data loss warnings', () => {
    describe('Return to task list', () => {
        it('locks focus to data loss warning dialog', () => {
            cy.visit('/fixtures?redirect=/choose-attorneys&progress=provideYourDetails');

            cy.get('#f-first-names').type('John');
            cy.contains('a', 'Return to task list').click()

            cy.get('#dialog').should('be.visible')
            cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Back to page')

            cy.realPress("Tab")
            cy.focused().should('contain', 'Continue without saving')

            cy.realPress("Tab")
            cy.focused().should('have.attr', 'id', 'dialog-title')

            cy.realPress("Tab")
            cy.focused().should('have.attr', 'id', 'dialog-description')

            cy.realPress("Tab")
            cy.focused().should('contain', 'Back to page')

            cy.contains('button', 'Back to page').click()

            cy.get('#dialog').should('not.be.visible')
            cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Return to task list')

            cy.contains('a', 'Return to task list').click()

            cy.get('#dialog').should('be.visible')
            cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Back to page')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('have.attr', 'id', 'dialog-description')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('have.attr', 'id', 'dialog-title')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('contain', 'Continue without saving')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('contain', 'Back to page')

            cy.realType('{esc}')

            cy.get('#dialog').should('not.be.visible')
            cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Return to task list')
        })
    })

    describe('Change language', () => {
        it('locks focus to data loss warning dialog', () => {
            cy.visit('/fixtures?redirect=/choose-attorneys&progress=provideYourDetails');

            cy.get('#f-first-names').type('John');
            cy.contains('a', 'Cymraeg').click()

            cy.get('#language-dialog').should('be.visible')
            cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Back to page')

            cy.realPress("Tab")
            cy.focused().should('contain', 'Continue without saving')

            cy.realPress("Tab")
            cy.focused().should('have.attr', 'id', 'language-dialog-title')

            cy.realPress("Tab")
            cy.focused().should('have.attr', 'id', 'language-dialog-description')

            cy.realPress("Tab")
            cy.focused().should('contain', 'Back to page')

            cy.contains('#language-dialog button', 'Back to page').click()

            cy.get('#dialog').should('not.be.visible')
            cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Cymraeg')

            cy.contains('a', 'Cymraeg').click()

            cy.get('#language-dialog').should('be.visible')
            cy.get('#dialog-overlay').should('not.have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Back to page')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('have.attr', 'id', 'language-dialog-description')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('have.attr', 'id', 'language-dialog-title')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('contain', 'Continue without saving')

            cy.realPress(["Shift", "Tab"])
            cy.focused().should('contain', 'Back to page')

            cy.realType('{esc}')

            cy.get('#language-dialog').should('not.be.visible')
            cy.get('#dialog-overlay').should('have.class', 'govuk-!-display-none')

            cy.focused().should('contain', 'Cymraeg')
        })
    })
})
