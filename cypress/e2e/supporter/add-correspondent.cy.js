import { AddressFormAssertions } from "../../support/e2e";

describe('Add correspondent', () => {
  beforeEach(() => {
    cy.visit('/fixtures/supporter?redirect=/dashboard&organisation=1&lpa=1');
  });

  it('allows none', () => {
    cy.contains('M-FAKE-').click();
    cy.contains('Go to task list').click();
    cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

    cy.checkA11yApp();
    cy.contains('label', 'No').click();
    cy.contains('button', 'Save and continue').click();
    cy.contains('li', 'Add a correspondent').should('contain', 'Completed');
  });

  it('allows without address', () => {
    cy.contains('M-FAKE-').click();
    cy.contains('Go to task list').click();
    cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

    cy.checkA11yApp();
    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Save and continue').click();

    cy.checkA11yApp();
    cy.get('#f-first-names').invoke('val', 'John');
    cy.get('#f-last-name').invoke('val', 'Smith');
    cy.get('#f-email').invoke('val', 'email@example.com');
    cy.contains('label', 'No').click();
    cy.contains('button', 'Save and continue').click();
    cy.contains('li', 'Add a correspondent').should('contain', 'Completed');
  });

  it('allows with address', () => {
    cy.contains('M-FAKE-').click();
    cy.contains('Go to task list').click();
    cy.contains('li', 'Add a correspondent').should('contain', 'Not started').click();

    cy.checkA11yApp();
    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Save and continue').click();

    cy.checkA11yApp();
    cy.get('#f-first-names').invoke('val', 'John');
    cy.get('#f-last-name').invoke('val', 'Smith');
    cy.get('#f-email').invoke('val', 'email@example.com');
    cy.contains('label', 'Yes').click();
    cy.contains('button', 'Save and continue').click();

    cy.contains('label', 'Enter a new address').click();
    cy.contains('button', 'Continue').click();
    AddressFormAssertions.assertCanAddAddressFromSelect()

    cy.contains('li', 'Add a correspondent').should('contain', 'Completed');
  });
});
