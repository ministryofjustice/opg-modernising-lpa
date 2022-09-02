describe('Task list', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list');
        cy.injectAxe();
    });

    it('shows tasks', () => {
        cy.contains('li', "Provide the donor's details").should('contain', 'Completed');
        cy.contains('li', 'Choose your attorneys').should('contain', 'Not started');
        cy.contains('li', 'Sign the LPA').should('contain', 'Cannot start yet');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });
    });
});
