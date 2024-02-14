import { TestMobile } from "../../support/e2e";

describe('Your independent witness mobile', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/your-independent-witness-mobile');
  });

  it('can be submitted', () => {
    cy.get('#f-mobile').type(TestMobile);

    cy.checkA11yApp({ rules: { 'aria-allowed-attr': { enabled: false } } });

    cy.contains('button', 'Save and continue').click();
    cy.url().should('contain', '/your-independent-witness-address');
  });

  it('errors when empty', () => {
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Enter a UK mobile number');
    });

    cy.contains('[for=f-mobile] + div + .govuk-error-message', 'Enter a UK mobile number');
  });

  it('errors when invalid mobile number', () => {
    cy.get('#f-mobile').type('not-a-number');
    cy.contains('button', 'Save and continue').click();

    cy.contains('[for=f-mobile] + div + .govuk-error-message', 'Enter a mobile number in the correct format');
  });

  it('errors when invalid non uk mobile number', () => {
    cy.get('#f-has-non-uk-mobile').check({ force: true });
    cy.get('#f-non-uk-mobile').type('not-a-number', { force: true });
    cy.contains('button', 'Save and continue').click();

    cy.contains('[for=f-non-uk-mobile] + div + .govuk-error-message', 'Enter a mobile number in the correct format');
  });
});
