describe('Certificate provided', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/certificate-provided');
    });

    it('has a button to the dashboard', () => {
        cy.checkA11yApp();
        cy.contains('a', 'Go to your dashboard');
    });
})
