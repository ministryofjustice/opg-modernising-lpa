describe('Certificate provided', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provided&loginAs=certificate-provider');
    });

    it('has a button', () => {
        cy.checkA11yApp();
        cy.contains('button', 'Go to your dashboard');
    });
})
