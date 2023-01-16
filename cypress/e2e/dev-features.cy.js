describe('Dev features', () => {

    it('Show translation keys', () => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used?showTransKeys=1&withAttorney=1');
        cy.get('h1').should(
            "have.text",
            "{When can your attorneys use your LPA} [whenCanTheLpaBeUsed]"
        );
        cy.contains('a', 'Toggle translation keys').click();

        cy.url().should('contain', '/when-can-the-lpa-be-used');
        cy.get('h1').should(
            "have.text",
            "When can your attorneys use your LPA"
        );
    });
});
