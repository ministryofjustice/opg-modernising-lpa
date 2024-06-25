describe('Enter voucher', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/enter-voucher&progress=payForTheLpa');
  });

  it('adds a voucher', () => {
    cy.checkA11yApp();

    cy.get('#f-first-names').type('Shopping');
    cy.get('#f-last-name').type('Voucher');
    cy.get('#f-email').type('voucher@example.com');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/task-list');
  });

  it('errors when empty', () => {
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Enter first names');
      cy.contains('Enter last name');
      cy.contains('Enter email address');
    });

    cy.contains('[for=f-first-names] + .govuk-error-message', 'Enter first names');
    cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
    cy.contains('[for=f-email] + .govuk-error-message', 'Enter email address');
  });

  it('errors when invalid', () => {
    cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
    cy.get('#f-last-name').invoke('val', 'b'.repeat(62));
    cy.get('#f-email').type('voucher');
    cy.contains('button', 'Save and continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('First names must be 53 characters or less');
      cy.contains('Last name must be 61 characters or less');
      cy.contains('Email address must be in the correct format, like name@example.com');
    });

    cy.contains('[for=f-first-names] + .govuk-error-message', 'First names must be 53 characters or less');
    cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
    cy.contains('[for=f-email] + .govuk-error-message', 'Email address must be in the correct format, like name@example.com');
  });
});
