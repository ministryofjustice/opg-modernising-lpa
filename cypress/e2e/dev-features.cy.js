describe('Dev features', () => {

    it('Show translation keys', () => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used?showTransKeys=1&withAttorney=1');
        cy.contains('h1', '{When can your attorneys use your LPA} [whenCanTheLpaBeUsed]');
        cy.contains('a', 'Toggle translation keys').click();

        cy.url().should('contain', '/when-can-the-lpa-be-used');
        cy.contains('h1', 'When can your attorneys use your LPA');
    });
});
