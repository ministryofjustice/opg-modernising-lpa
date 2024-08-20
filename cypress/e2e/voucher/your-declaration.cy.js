describe('Confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?redirect=/sign-the-declaration&progress=confirmYourIdentity');
    });

    it('can be signed', () => {
        cy.checkA11yApp();
        cy.contains('label', 'To the best of my knowledge').click();
        cy.contains('button', 'Submit my signature').click();

        // TODO: this will change when the next ticket is picked up
        cy.url().should('contain', '/task-list');
    });
});
