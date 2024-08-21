describe('Confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?redirect=/sign-the-declaration&progress=confirmYourIdentity');
    });

    it('can be signed', () => {
        cy.checkA11yApp();
        cy.contains('label', 'To the best of my knowledge').click();
        cy.contains('button', 'Submit my signature').click();

        cy.url().should('contain', '/thank-you');
        cy.contains('a', 'Manage your LPAs').click();
        cy.contains('I’m vouching for someone').should('not.exist');
    });
});
