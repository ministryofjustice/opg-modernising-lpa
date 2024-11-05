describe('Choose someone to vouch for you', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/choose-someone-to-vouch-for-you&progress=payForTheLpa')
        cy.url().should('contain', '/choose-someone-to-vouch-for-you')
        cy.checkA11yApp()
    })

    it('errors when option not selected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/choose-someone-to-vouch-for-you')
        cy.checkA11yApp()

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes if you know someone and they have agreed to vouch for you');
        });

        cy.contains('.govuk-error-message', 'Select yes if you know someone and they have agreed to vouch for you');
    })
})
