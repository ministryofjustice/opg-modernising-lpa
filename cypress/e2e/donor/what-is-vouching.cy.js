describe('what is vouching', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/what-is-vouching&progress=payForTheLpa')
        cy.url().should('contain', '/what-is-vouching')
        cy.checkA11yApp()
    })

    it('errors when option not selected', () => {
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/what-is-vouching')
        cy.checkA11yApp()

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select yes if you know someone and they have agreed to vouch for you');
        });

        cy.contains('.govuk-error-message', 'Select yes if you know someone and they have agreed to vouch for you');
    })
})
