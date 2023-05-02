describe('Read the LPA', () => {
    it('displays the LPA details with actor specific content', () => {
        cy.visit('/testing-start?redirect=/attorney-read-the-lpa&completeLpa=1&withAttorney=1&asAttorney=1');

        cy.contains('dt', "When attorneys can use the LPA")
        cy.contains('dt', "Attorney names")
        cy.contains('dt', "Replacement attorney names")

        cy.contains('Continue').click();

        cy.url().should('contain', '/attorney-sign');
    });
});
