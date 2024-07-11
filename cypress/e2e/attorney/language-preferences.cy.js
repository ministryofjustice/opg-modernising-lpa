describe('Your preferred language', () => {
  beforeEach(() => {
    cy.visit('/fixtures/attorney?redirect=/your-preferred-language');
    cy.url().should('contain', '/your-preferred-language')
  });

  it('can choose a language contact preference', () => {
    cy.get('[name="language-preference"]').check('en', { force: true })

    cy.checkA11yApp();

    cy.contains('button', 'Save and continue').click()

    cy.url().should('contain', '/confirm-your-details')
  })

  it('errors when preference not selected', () => {
    cy.contains('button', 'Save and continue').click()
    cy.url().should('contain', '/your-preferred-language')

    cy.checkA11yApp();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select which language you would like us to use when we contact you');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select which language you would like us to use when we contact you');
  })
})
