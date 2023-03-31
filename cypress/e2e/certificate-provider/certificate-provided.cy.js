describe('Certificate provided', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provided&completeLpa=1&asCertificateProvider=1');
    });

    it('has a button', () => {
        cy.checkA11yApp();
        cy.contains('button', 'Go to your dashboard');
    });
})
