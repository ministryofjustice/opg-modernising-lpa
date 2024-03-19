const { TestEmail } = require("../../support/e2e");

describe('Donor access', () => {
  describe('invite can be', () => {
    beforeEach('', () => {
      cy.visit('/fixtures/supporter?redirect=/dashboard&organisation=1&lpa=1');
    })

    it('sent and recalled', () => {
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
  })

  describe('access can be', () => {
    beforeEach('', () => {
      cy.visit('/fixtures/supporter?redirect=/dashboard&organisation=1&accessCode=abcdef123456&linkDonor=1');
    })

    it('removed', () => {
      cy.contains('a', 'M-FAKE-').click()
      cy.contains('Donor access').click();

      cy.contains('dt', 'Status').parent().contains('Linked')
      cy.contains('button', 'Remove access').click()

      cy.url().should('contain', '/view-lpa');
      cy.checkA11yApp();

      cy.contains(`You removed access to this LPA for ${TestEmail}.`)

      cy.contains('Donor access').click();

      cy.contains('dt', 'Status').parent().should('not.contain', 'Linked')
    })
  })
});
