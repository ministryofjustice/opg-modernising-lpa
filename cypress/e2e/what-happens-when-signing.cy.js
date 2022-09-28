describe('What happens when signing', () => {
    it('has a continue button', () => {
        cy.visit('/testing-start?redirect=/what-happens-when-signing');

        cy.injectAxe();
        cy.checkA11y(null, {
            rules: { region: { enabled: false } },
        });

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
    });
});
