describe('High level navigation', () => {
    describe('back link', () => {
        it('navigates to the previous page', () => {
            cy.visit('/fixtures?redirect=/task-list&progress=provideYourDetails')
            cy.contains('a', 'Choose your attorneys').click()

            cy.url().should('contain', '/choose-attorneys-guidance')
            cy.contains('a', 'Continue').click()

            cy.url().should('contain', '/enter-attorney')
            cy.contains('a', 'Back').click()

            cy.url().should('contain', '/choose-attorneys-guidance')
        })
    })
})
