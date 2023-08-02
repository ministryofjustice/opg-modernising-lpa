describe('Task list', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list&lpa.complete=1&asCertificateProvider=1&loginAs=certificate-provider');
    });

    it('shows tasks', () => {
        cy.checkA11yApp();

        cy.contains('li', 'Confirm your details').should('contain', 'Not started');
        cy.contains('li', 'Confirm your identity').should('contain', 'Not started');
        cy.contains('li', 'Read the LPA').should('contain', 'Not started');
        cy.contains('li', 'Provide the certificate for this LPA').should('contain', 'Cannot start yet');
    });
});
