describe('Enter reference number', () => {
    beforeEach(() => {
        cy.visit('/testing-start?asAttorney=1&redirect=/attorney-enter-reference-number');
    });

    it('is a placeholder page', () => {
        cy.url().should('contain', '/attorney-enter-reference-number');
    });
});
