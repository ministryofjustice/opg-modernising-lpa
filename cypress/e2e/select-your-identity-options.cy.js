describe('Select your identity options', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/select-your-identity-options');
    });

    it('can be submitted', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'A passport').click();
        cy.contains('label', 'A driving licence').click();
        cy.contains('label', 'A utility bill').click();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-chosen-identity-options');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('using your passport and driving licence');

        cy.contains('button', 'Continue');
    });
});
