describe('Read the LPA', () => {
    it('displays the LPA details with actor specific content', () => {
        cy.visit('/testing-start?redirect=/read-the-lpa&lpa.complete=1&attorneyProvided=1&loginAs=attorney');

        cy.contains('dt', "When attorneys can use the LPA")
        cy.contains('dt', "Attorney names")
        cy.contains('dt', "Replacement attorney names")

        cy.contains('Continue').click();

        cy.url().should('contain', '/legal-rights-and-responsibilities');
    });
});
