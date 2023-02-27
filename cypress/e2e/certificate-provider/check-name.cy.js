describe('Check your name', () => {
    it('can confirm name matches', () => {
        cy.visit('/testing-start?redirect=/certificate-provider-check-name&completeLpa=1&asCertificateProvider=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-is-name-correct').check('yes');
        cy.contains('Continue').click();

        cy.url().should('contain', '/certificate-provider-details');
    });
});
