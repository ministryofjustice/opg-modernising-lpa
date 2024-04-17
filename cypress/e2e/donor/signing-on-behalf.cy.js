describe('Signing on behalf of the donor', () => {
  beforeEach(() => {
    cy.visit('/fixtures?redirect=/how-to-confirm-your-identity-and-sign&progress=payForTheLpa&donor=cannot-sign');

    cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
    cy.contains('h1', 'How to confirm your identity and sign the LPA');
    cy.contains('a', 'Continue').click();
    cy.contains('a', 'Continue').click();

    cy.contains('label', 'Sam Smith (donor)').click();
    cy.contains('button', 'Continue').click();
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/read-your-lpa');
    cy.contains('a', 'Continue').click();

    cy.url().should('contain', '/your-legal-rights-and-responsibilities');
    cy.contains('a', 'Continue to signing page').click();
  });

  it('can be completed', () => {
    cy.url().should('contain', '/sign-the-lpa-on-behalf');
    cy.checkA11yApp();

    cy.contains('h1', "Sign your LPA");
    cy.contains('label', 'Sam Smith wants to sign this LPA as a deed').click();
    cy.contains('label', 'Sam Smith wants to apply to register this LPA').click();
    cy.contains('button', 'Submit signature').click();

    cy.url().should('contain', '/witnessing-your-signature');
    cy.checkA11yApp();
    cy.contains('your independent witness, Indie Irwin');
    cy.contains('your certificate provider, Charlie Cooper');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/witnessing-as-independent-witness');
    cy.checkA11yApp();
    cy.get('#f-witness-code').type('1234');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/witnessing-as-certificate-provider');
    cy.checkA11yApp();
    cy.get('#f-witness-code').type('1234');
    cy.contains('button', 'Continue').click();

    cy.url().should('contain', '/you-have-submitted-your-lpa');
  });
});
