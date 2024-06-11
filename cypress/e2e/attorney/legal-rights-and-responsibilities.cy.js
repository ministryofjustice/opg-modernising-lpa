describe('Legal rights and responsibilities', () => {
    it('can continue to next page', () => {
        cy.visit('/fixtures/attorney?redirect=/legal-rights-and-responsibilities&progress=readTheLpa');

        cy.contains('h1', "Your legal rights and responsibilities")

        cy.contains('Continue').click();

        cy.url().should('contain', '/what-happens-when-you-sign');
    });
});
