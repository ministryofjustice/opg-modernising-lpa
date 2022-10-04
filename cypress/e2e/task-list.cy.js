describe('Task list', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list');
        cy.injectAxe();
    });

    it('shows tasks', () => {
        cy.contains('li', "Provide the donor's details").should('contain', 'Not started');
        cy.contains('li', 'Register the LPA').should('contain', 'Cannot start yet');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });
    });
});
