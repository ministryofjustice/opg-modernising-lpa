describe('Enter group name', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?redirect=/enter-the-name-of-your-organisation-or-company');
  });

  it('can be started', () => {
    cy.checkA11yApp();
    cy.get('#f-name').type('My name' + Math.random());
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/organisation-or-company-created');
  });
});
