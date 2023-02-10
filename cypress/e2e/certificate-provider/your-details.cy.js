describe('Your details', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-your-details&completeLpa=1&asCertificateProvider=1');
    });

    it('can be completed', () => {
        cy.contains('Donor is Jose Smith');
        cy.contains('Certificate provider is Barbara Smith');
    });
});
