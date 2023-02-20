describe('Read the LPA', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-read-the-lpa&completeLpa=1&asCertificateProvider=1');
    });

    it('can be read', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('a', 'Continue').click();
        cy.url().should('contain', '/being-a-certificate-provider');
    });
});
