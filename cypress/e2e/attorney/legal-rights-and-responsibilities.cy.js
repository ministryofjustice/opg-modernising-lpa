describe('Legal rights and responsibilities', () => {
    it('can continue to next page', () => {
        cy.visit('/testing-start?redirect=/attorney-legal-rights-and-responsibilities&completeLpa=1&withAttorney=1&asAttorney=1');

        cy.contains('h1', "Your legal rights and responsibilities")

        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-what-happens-when-you-sign');
    });
});
