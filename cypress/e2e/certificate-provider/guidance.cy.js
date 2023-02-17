describe('Guidance', () => {
    describe('when the LPA is signed', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/being-a-certificate-provider&completeLpa=1&asCertificateProvider=1');
        });

        it('goes to the next step', () => {
            cy.contains('Continue').click();
            cy.url().should('contain', '/certificate-provider-next');
        });
    });

    describe('when the LPA is not yet signed', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/being-a-certificate-provider&withCP=1&withDonorDetails=1&asCertificateProvider=1');
        });

        it('goes to a confirmation page', () => {
            cy.contains('Continue').click();
            cy.url().should('contain', '/certificate-provider-confirmation');
        });
    });
});
