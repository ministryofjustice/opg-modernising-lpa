describe('Check your details', () => {
  it('shows my details', () => {
    cy.visit('/fixtures?redirect=/check-your-details&progress=confirmYourIdentity&idStatus=donor:insufficient-evidence&voucher=1');

    cy.checkA11yApp();
    cy.contains('Sam Smith');
    cy.contains('2 January 2000');
    cy.contains('1 RICHMOND PLACE');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/we-have-contacted-voucher');
    cy.checkA11yApp();

    cy.contains('a', 'Continue').click();

    cy.url().should('contain', '/read-your-lpa');
    cy.checkA11yApp();
  });

  it('tells me about a pending payment', () => {
    cy.visit('/fixtures?redirect=/check-your-details&progress=payForTheLpa&feeType=NoFee&paymentTaskProgress=Pending');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/we-have-received-voucher-details');
    cy.checkA11yApp();
    cy.contains('no fee (exemption)');
  });
});
