describe('Legal rights and responsibilities', () => {
    it('can continue to next page', () => {
        cy.visit('/testing-start?redirect=/legal-rights-and-responsibilities&lpa.complete=1&lpa.attorneys=1&lpa.signedByDonor=1&attorneyProvided=1&loginAs=attorney');

        cy.contains('h1', "Your legal rights and responsibilities")

        cy.contains('Continue').click();

        cy.url().should('contain', '/what-happens-when-you-sign');
    });
});
