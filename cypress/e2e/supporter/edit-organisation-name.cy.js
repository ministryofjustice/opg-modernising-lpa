describe('Edit organisation name', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/organisation-details/edit-organisation-name');
  });

  it('can be started', () => {
    cy.checkA11yApp();
    cy.get('#f-name').clear().type('My organisation');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/manage-organisation/organisation-details');
    cy.checkA11yApp();
    cy.contains('Your organisation name has been saved.');
    cy.contains('.govuk-summary-list', 'My organisation');
  });
});
