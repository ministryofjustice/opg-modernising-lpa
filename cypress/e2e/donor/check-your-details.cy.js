describe('Check your details', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/check-your-details&progress=payForTheLpa');
  });

  it('shows my details', () => {
    cy.checkA11yApp();
    cy.contains('Sam Smith');
    cy.contains('2 January 2000');
    cy.contains('1 RICHMOND PLACE');
    cy.contains('a', 'Continue').click();

    cy.url().should('contain', '/task-list');
  });
});
