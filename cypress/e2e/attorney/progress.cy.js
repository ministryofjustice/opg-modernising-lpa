describe('Progress', () => {
    it('when nothing completed', () => {
        cy.visit('/fixtures/attorney?redirect=/progress');
        cy.checkA11yApp();

        cy.contains('li', 'LPA signed by you Not completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
    });

    it('when signed', () => {
        cy.visit('/fixtures/attorney?redirect=/progress&progress=signedByAttorney');
        cy.checkA11yApp();

        cy.contains('li', 'LPA signed by you Completed');
        cy.contains('li', 'LPA signed by all attorneys Not completed');
    });

    it('when all signed', () => {
        cy.visit('/fixtures/attorney?redirect=/progress&progress=signedByAllAttorneys');
        cy.checkA11yApp();

        cy.contains('li', 'LPA signed by you Completed');
        cy.contains('li', 'LPA signed by all attorneys Completed');
    });
});
