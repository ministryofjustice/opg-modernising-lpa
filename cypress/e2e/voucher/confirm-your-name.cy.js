describe('Confirm your name', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?redirect=/confirm-your-name');
    });

    it('shows my name', () => {
        cy.checkA11yApp();

        cy.contains('Vivian');
        cy.contains('Vaughn');
    });

    it('can update my name', () => {
        cy.contains('div', 'Vivian').contains('a', 'Change').click();

        cy.url().should('contain', '/your-name')
        cy.checkA11yApp();
        cy.get('#f-first-names').clear().type('Barry');
        cy.contains('button', 'Save and continue').click();

        cy.url().should('contain', '/confirm-your-name')
        cy.contains('Barry');
    });
});
