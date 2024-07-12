describe('Your preferred language', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/your-preferred-language');
    cy.url().should('contain', '/your-preferred-language')
  });

  it('can choose language preferences', () => {
    cy.get('[name="contact-language"]').check('en', { force: true })
    cy.get('[name="lpa-language"]').check('en', { force: true })

    cy.checkA11yApp();

    cy.contains('button', 'Save and continue').click()

    cy.url().should('contain', '/lpa-type')
  })

  it('errors when preference not selected', () => {
    cy.contains('button', 'Save and continue').click()
    cy.url().should('contain', '/your-preferred-language')

    cy.checkA11yApp();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select which language you would like us to use when we contact you');
      cy.contains('Select the language in which you would like your LPA registered');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select which language you would like us to use when we contact you');
    cy.contains('.govuk-fieldset .govuk-error-message', 'Select the language in which you would like your LPA registered');
  })
})
