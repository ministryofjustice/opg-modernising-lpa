describe('Do you want replacement attorneys', () => {
    it('can be submitted - acting jointly', () => {
        cy.visit('/testing-start?redirect=/want-replacement-attorneys&howAttorneysAct=jointly');
        cy.injectAxe();

        cy.contains('Replacement attorneys are an important backup when attorneys are appointed to act jointly.')

        cy.get('#f-want').check()

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });

    it('can be submitted - acting jointly for some and severally for others', () => {
        cy.visit('/testing-start?redirect=/want-replacement-attorneys&howAttorneysAct=mixed');
        cy.injectAxe();

        cy.contains('The donor appointed their attorneys to act jointly for some decisions, and jointly and severally for others.')

        cy.get('#f-want').check()

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys');
    });
});
