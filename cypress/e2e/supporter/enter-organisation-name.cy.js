describe('Enter organisation name', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?redirect=/enter-your-name');
    });

    it('can be started', () => {
        cy.checkA11yApp();
        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Smith');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/enter-the-name-of-your-organisation-or-company');
        cy.checkA11yApp();

        cy.get('#f-name').invoke('val', 'My name' + Math.random());
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/organisation-or-company-created');
    });
});
