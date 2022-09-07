describe('When can the LPA be used', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/when-can-the-lpa-be-used');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-when').check('when-registered');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/restrictions');
    });
});
