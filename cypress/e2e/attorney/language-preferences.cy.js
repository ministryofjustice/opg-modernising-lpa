describe('Your preferred language', () => {
    it('can choose a language contact preference', () => {
        cy.visit('/fixtures/attorney?redirect=/your-preferred-language');

        cy.get('#f-language-preference').check('en')

        cy.checkA11yApp();

        cy.contains('button', 'Save and continue').click()

        cy.url().should('contain', '/confirm-your-details')
    })

    it('errors when preference not selected', () => {
        cy.visit('/fixtures/attorney?redirect=/your-preferred-language');

        cy.contains('button', 'Save and continue').click()
        cy.url().should('contain', '/your-preferred-language')

        cy.checkA11yApp();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select your preferred language');
        });

        cy.contains('.govuk-fieldset .govuk-error-message', 'Select your preferred language');
    })
})
