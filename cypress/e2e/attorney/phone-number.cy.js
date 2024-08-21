import { TestMobile } from "../../support/e2e";

describe('Phone number', () => {
  beforeEach(() => {
    cy.visit('/fixtures/attorney?redirect=/phone-number');
  });

  it('can be completed', () => {
    cy.checkA11yApp();

    cy.get('#f-phone').type(TestMobile);

    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/your-preferred-language');
  });

  it('can be empty', () => {
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/your-preferred-language');
  });

  it('errors when not a phone number', () => {
    cy.get('#f-phone').type('not a mobile');
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Phone must be a phone number');
    });

    cy.contains('[for=f-phone] ~ .govuk-error-message', 'Phone must be a phone number');
  });
});
