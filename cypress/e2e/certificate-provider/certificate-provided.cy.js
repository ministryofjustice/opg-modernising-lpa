describe('Certificate provided', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provided&loginAs=certificate-provider');
    });

    it('has a button to the dashboard', () => {
        cy.checkA11yApp();
        cy.contains('a', 'Go to your dashboard');
    });
})
