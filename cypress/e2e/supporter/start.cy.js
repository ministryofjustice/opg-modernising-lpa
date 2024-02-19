describe('Start', () => {
  beforeEach(() => {
    cy.visit('/supporter-start');
  });

  it('can be started', () => {
    cy.checkA11yApp();
    cy.contains("Help someone to make a lasting power of attorney");
    cy.contains('a', 'Start').click();

    cy.checkA11yApp();
    cy.contains("Signing in with GOV.UK One Login");
    cy.contains('a', 'Continue to GOV.UK One Login').click();

    if (Cypress.config().baseUrl.includes('localhost')) {
      cy.url().should('contain', '/authorize')
      cy.contains('button', 'Sign in').click();

      cy.url().should('contain', '/enter-the-name-of-your-organisation-or-company')
    } else {
      cy.origin('https://signin.integration.account.gov.uk', () => {
        cy.url().should('contain', '/')
      })
    }
  });
});
