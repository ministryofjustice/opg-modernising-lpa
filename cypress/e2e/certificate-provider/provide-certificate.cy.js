describe('Provide the certificate', () => {
  beforeEach(() => {
    cy.visit('/fixtures/certificate-provider?redirect=/provide-certificate&progress=signedByDonor');
  });

  it('can provide the certificate', () => {
    cy.checkA11yApp();

    cy.get('#f-agree-to-statement').check({ force: true })

    cy.contains('button', 'Submit signature').click();
    cy.url().should('contain', '/certificate-provided');
  });

  it("errors when not selected", () => {
    cy.contains('button', 'Submit signature').click();

    cy.get('.govuk-error-summary').within(() => {
      cy.contains('Select the box to sign as the certificate provider');
    });

    cy.contains('.govuk-form-group .govuk-error-message', 'Select the box to sign as the certificate provider');
  })
});
