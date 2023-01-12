describe('Dev features', () => {

    it('Show translation keys', () => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used?showTransKeys&withAttorney=1');
        cy.contains('h1', '{When can your attorneys use your LPA} [whenCanTheLpaBeUsed]');

        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used&withAttorney=1');
        cy.contains('h1', 'When can your attorneys use your LPA');
    });
});
