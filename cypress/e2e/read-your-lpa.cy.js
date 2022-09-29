describe('Check the LPA', () => {
    it("submits the completed LPA", () => {
        cy.visit('/testing-start?redirect=/read-your-lpa');

        cy.contains('h1', "Read and sign your LPA");

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('h2', "LPA decisions");

        cy.contains('h2', "People named on the LPA");
        cy.contains('h3', "Donor");
        cy.contains('h3', "Attorneys");
        cy.contains('h3', "Replacement attorney");

        cy.get('#f-checked').check();
        cy.get('#f-confirm').check();
        cy.get('#f-signature').type('1234');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');
    })
});
