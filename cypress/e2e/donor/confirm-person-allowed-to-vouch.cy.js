describe('Enter voucher', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/enter-voucher&progress=payForTheLpa');

    cy.get('#f-first-names').type('Sam');
    cy.get('#f-last-name').type('Smith');
    cy.get('#f-email').type('voucher@example.com');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/confirm-person-allowed-to-vouch');
  });

  it('can confirm', () => {
    cy.checkA11yApp();

    cy.contains('You have entered a name which matches your name, Sam Smith.');
    cy.get('input[name=yes-no]').check('yes');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/check-your-details');
  });

  it('can select another', () => {
    cy.get('input[name=yes-no]').check('no');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/enter-voucher');
    cy.get('#f-first-names').should('have.value', '');
    cy.get('#f-last-name').should('have.value', '');
    cy.get('#f-email').should('have.value', '');
  });

  it('errors when not selected', () => {
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/confirm-person-allowed-to-vouch');
    cy.checkA11yApp();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select yes, if the person is allowed to vouch for you');
    });

    cy.contains('.govuk-fieldset .govuk-error-message', 'Select yes, if the person is allowed to vouch for you');
  });
});
