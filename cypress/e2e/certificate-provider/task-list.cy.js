describe('Task list', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=signedByDonor');
    });

    it('shows tasks', () => {
        cy.checkA11yApp();

        cy.contains('li', 'Confirm your details').should('contain', 'Not started');
        cy.contains('li', 'Confirm your identity').should('contain', 'Not started');
        cy.contains('li', 'Provide the certificate for this LPA').should('contain', 'Cannot start yet');
    });
});
