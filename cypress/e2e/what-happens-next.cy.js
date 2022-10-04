describe('What happens next', () => {
    it('has a continue button', () => {
        cy.visit('/testing-start?redirect=/what-happens-next');

        cy.injectAxe();
        cy.checkA11y(null, {
            rules: { region: { enabled: false } },
        });

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/about-payment');
    });
});
