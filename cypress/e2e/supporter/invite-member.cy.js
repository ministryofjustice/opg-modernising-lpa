const { TestEmail } = require("../../support/e2e");

describe('Invite member', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?organisation=1&redirect=/invite-member');
  });

  it('can be started', () => {
    cy.checkA11yApp();
    cy.get('#f-email').type(TestEmail);
    cy.contains('button', 'Send invite').click();

    cy.url().should('contain', '/invite-member-confirmation');
    cy.checkA11yApp();
    cy.contains(TestEmail);
  });
});
