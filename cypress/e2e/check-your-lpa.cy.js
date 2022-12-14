describe('Check the LPA', () => {
    it("can submit the completed LPA", () => {
        cy.visit('/testing-start?redirect=/check-your-lpa&withIncompleteAttorneys=1');

        cy.contains('h1', "Check your LPA")

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('h2', "LPA decisions")

        cy.contains('h2', "People named on the LPA")
        cy.contains('h3', "Donor")
        cy.contains('h3', "Certificate provider")
        cy.contains('h3', "Attorneys")
        cy.contains('h3', "Replacement attorney")

        cy.get('#f-checked').check()
        cy.get('#f-happy').check()

        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/task-list');
    })
});
