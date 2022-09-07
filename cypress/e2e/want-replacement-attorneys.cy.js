describe('Choose attorneys', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/want-replacement-attorneys');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-want').check()

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/when-can-the-lpa-be-used');
    });
});
