describe('Progress', () => {
    it('when nothing completed', () => {
        cy.visit('/fixtures/attorney?redirect=/progress');
        cy.checkA11yApp();

        cy.contains('li', 'You’ve signed the LPA In progress');
        cy.contains('li', 'All attorneys have signed the LPA Not started');
    });

    it('when signed', () => {
        cy.visit('/fixtures/attorney?redirect=/progress&progress=signedByAttorney');
        cy.checkA11yApp();

        cy.contains('li', 'You’ve signed the LPA Completed');
        cy.contains('li', 'All attorneys have signed the LPA In progress');
    });

    it('when all signed', () => {
        cy.visit('/fixtures/attorney?redirect=/progress&progress=signedByAllAttorneys');
        cy.checkA11yApp();

        cy.contains('li', 'You’ve signed the LPA Completed');
        cy.contains('li', 'All attorneys have signed the LPA Completed');
    });
});
