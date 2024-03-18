const { TestEmail } = require("../../support/e2e");

describe('Donor access', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?redirect=/dashboard&organisation=1&lpa=1');
  });

  it('shows donor access page', () => {
    cy.contains('a', 'M-FAKE-').click()
    cy.contains('Donor access').click();

    cy.checkA11yApp();
    cy.contains('M-FAKE-');
    cy.get('#f-email').should('have.value', TestEmail);

    cy.contains('button', 'Send invite').click();

    cy.url().should('contain', '/view-lpa');
    cy.contains(`You sent an invite to ${TestEmail}`);

    cy.contains('Donor access').click();
    cy.get('#f-email').should('not.exist');
    cy.contains('Pending');

    cy.contains('button', 'Recall invite').click();

    cy.url().should('contain', '/view-lpa');
    cy.contains(`You recalled the invite to this LPA for ${TestEmail}.`)

    cy.contains('Donor access').click();
    cy.get('#f-email');
  });
});
