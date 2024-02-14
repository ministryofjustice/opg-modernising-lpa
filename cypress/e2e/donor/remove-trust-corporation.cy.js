describe('Remove trust corporation', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/choose-attorneys-summary&progress=chooseYourAttorneys&attorneys=trust-corporation');
  });

  it('can be removed', () => {
    cy.contains('You have added 3 attorneys');
    cy.contains('a', 'Remove trust corporation').click();

    cy.contains('Are you sure you want to remove First Choice Trust Corporation Ltd.?');
    cy.get('input[name="yes-no"]').check('yes', { force: true });
    cy.contains('button', 'Continue').click();
    cy.contains('You have added 2 attorneys');
    cy.contains('trust corporation').should('not.exist');
  });

  it('errors when not selected', () => {
    cy.contains('a', 'Remove trust corporation').click();

    cy.contains('button', 'Continue').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select yes to remove the trust corporation');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes to remove the trust corporation');
  });
});
