describe('Task list', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list&withDonorDetails=1');
    });

    it('shows tasks', () => {
        cy.contains('li', "Provide your details").should('contain', 'Completed');
        cy.contains('li', "Choose your attorneys").should('contain', 'Not started');
        cy.contains('li', "Pay for the LPA").should('contain', 'Cannot start yet');
        cy.contains('li', 'Confirm your identity and sign the LPA').should('contain', 'Cannot start yet');

        cy.checkA11yApp();
    });
});
