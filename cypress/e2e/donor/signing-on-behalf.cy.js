describe('Signing on behalf of the donor', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/your-legal-rights-and-responsibilities&lpa.paid=1&lpa.certificateProvider=1&lpa.yourDetails=1&lpa.cannotSign=1');
    });

    it('can be completed', () => {
        cy.url().should('contain', '/your-legal-rights-and-responsibilities');
        cy.checkA11yApp();
        cy.contains('a', 'Continue to signing page').click();

        cy.url().should('contain', '/sign-the-lpa-on-behalf');
        cy.checkA11yApp();

        cy.contains('h1', "Sign the LPA on behalf of Sam Smith");
        cy.contains('label', 'Sam Smith wants to sign this LPA as a deed').click();
        cy.contains('label', 'Sam Smith wants to apply to register this LPA').click();
        cy.contains('button', 'Submit signature').click();
    });
});
