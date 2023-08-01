describe('Who is eligible', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/certificate-provider-who-is-eligible&loginAs=certificate-provider');
    });

    it('can continue', () => {
        cy.checkA11yApp();

        cy.contains('Continue').click();

        cy.url().should('contain', '/enter-date-of-birth')
    });
});
