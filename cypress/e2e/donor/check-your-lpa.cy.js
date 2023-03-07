describe('Check the LPA', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/check-your-lpa&withDonorDetails=1&withCP=1&withAttorney=1&withReplacementAttorneys=1&whenCanBeUsedComplete=1&withRestrictions=1&withPeopleToNotify=1');
    });

    it("can submit the completed LPA", () => {
        cy.contains('h1', "Check your LPA")

        cy.injectAxe();
        cy.checkA11yApp();

        cy.contains('h2', "LPA decisions")

        cy.contains('h2', "People named on the LPA")
        cy.contains('h3', "Donor")
        cy.contains('h3', "Certificate provider")
        cy.contains('h3', "Attorneys")

        cy.get('#f-checked').check()
        cy.get('#f-happy').check()

        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/task-list');
    });

    it("errors when not selected", () => {
        cy.contains('button', 'Confirm').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select that you have checked your LPA and don’t wish to make changes');
            cy.contains('Select that you are happy to share your LPA with your certificate provider');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Select that you have checked your LPA and don’t wish to make changes');
        cy.contains('.govuk-form-group .govuk-error-message', 'Select that you are happy to share your LPA with your certificate provider');
    })
});
