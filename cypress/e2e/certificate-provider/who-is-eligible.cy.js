describe('Who is eligible', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/certificate-provider-who-is-eligible');
    });

    it('can continue', () => {
        cy.checkA11yApp();

        cy.contains('Continue').click();

        cy.url().should('contain', '/enter-date-of-birth')
    });
});
