describe('How to sign', () => {
    it('has a continue button', () => {
        cy.visit('/testing-start?redirect=/how-to-sign');

        cy.injectAxe();
        cy.checkA11y(null, {
            rules: { region: { enabled: false } },
        });

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/read-your-lpa');
    });
});
