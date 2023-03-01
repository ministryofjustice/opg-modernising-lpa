describe('Read the LPA', () => {
    describe('when the LPA is signed', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/certificate-provider-read-the-lpa&completeLpa=1&asCertificateProvider=1');
        });

        it('goes to the next step', () => {
            cy.injectAxe();
            cy.checkA11y(null, { rules: { region: { enabled: false } } });

            cy.contains('Continue').click();
            cy.url().should('contain', '/provide-certificate');
        });
    });

    describe('when the LPA is not yet signed', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/certificate-provider-read-the-lpa&withCP=1&withDonorDetails=1&asCertificateProvider=1');
        });

        it('goes to a guidance page', () => {
            cy.contains('Continue').click();
            cy.url().should('contain', '/certificate-provider-what-happens-next');
        });
    });
});
