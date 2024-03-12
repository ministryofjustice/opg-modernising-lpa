describe('Suspended', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?redirect=/dashboard&organisation=1&suspended=1');
  });

  it('does not allow access to organisation', () => {
    cy.checkA11yApp();

    cy.contains('Access suspended');
    cy.contains('Dashboard').should('not.exist');
    cy.contains('Manage organisation').should('not.exist');
  });
});
