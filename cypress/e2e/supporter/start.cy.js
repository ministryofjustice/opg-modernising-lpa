describe('Start', () => {
  beforeEach(() => {
    cy.visit('/supporter-start');
  });

  it('can be started', () => {
    cy.contains("Help someone to make a lasting power of attorney");
    cy.contains('a', 'Start').click();

    if (Cypress.config().baseUrl.includes('localhost')) {
      cy.url().should('contain', '/authorize')
    } else {
      cy.origin('https://signin.integration.account.gov.uk', () => {
        cy.url().should('contain', '/')
      })
    }
  });
});
