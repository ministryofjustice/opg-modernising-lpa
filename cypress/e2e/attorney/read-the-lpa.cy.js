describe('Read the LPA', () => {
    it('displays the LPA details', () => {
        cy.visit('/testing-start?redirect=/attorney-read-the-lpa&completeLpa=1&withAttorney=1&asAttorney=1');

        cy.contains('h2', "LPA decisions")

        cy.contains('h2', "People named on the LPA")
        cy.contains('h3', "Donor")
        cy.contains('h3', "Certificate provider")
        cy.contains('h3', "Attorneys")
        cy.contains('h3', "Replacement Attorneys")

        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-next-page');
    });
});
