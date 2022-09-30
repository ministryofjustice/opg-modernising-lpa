describe('Select your identity options', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/select-your-identity-options');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.contains('label', 'A passport').click();
        cy.contains('label', 'A driving licence').click();
        cy.contains('label', 'A utility bill').click();

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
    });
});
