describe('Dev features disabled', () => {

    it('Show translation keys', () => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used?showTransKeys=1&withAttorney=1');
        cy.get('h1').should(
            "have.text",
            "When can your attorneys use your LPA"
        );
        cy.contains('a', 'Toggle translation keys').should('not.exist');
    });
});
