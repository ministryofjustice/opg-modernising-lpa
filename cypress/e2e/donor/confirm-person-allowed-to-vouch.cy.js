describe('Enter voucher', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/enter-voucher&progress=payForTheLpa');
  });

  it('can confirm', () => {
    cy.get('#f-first-names').type('Sam');
    cy.get('#f-last-name').type('Smith');
    cy.get('#f-email').type('voucher@example.com');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/confirm-person-allowed-to-vouch');
    cy.checkA11yApp();

    cy.contains('You have entered a name which matches your name, Sam Smith.');
    cy.get('input[name=yes-no]').check('yes');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/check-your-details');
  });

  it('can select another', () => {
    cy.get('#f-first-names').type('Sam');
    cy.get('#f-last-name').type('Smith');
    cy.get('#f-email').type('voucher@example.com');
    cy.contains('button', 'Save and continue').click();

    cy.get('input[name=yes-no]').check('no');
    cy.contains('button', 'Save and continue').click();

    cy.url().should('contain', '/enter-voucher');
    cy.get('#f-first-names').should('have.value', '');
    cy.get('#f-last-name').should('have.value', '');
    cy.get('#f-email').should('have.value', '');
  });
});
