describe('Who is eligible', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-who-is-eligible&completeLpa=1&asCertificateProvider=1');
    });

    it('can continue', () => {
        cy.checkA11yApp();

        cy.contains('Continue').click();
    });
});
