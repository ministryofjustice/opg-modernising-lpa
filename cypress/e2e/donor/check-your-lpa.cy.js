describe('Check the LPA', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/check-your-lpa&lpa.yourDetails=1&lpa.certificateProvider=1&lpa.attorneys=1&lpa.replacementAttorneys=2&lpa.chooseWhenCanBeUsed=1&lpa.restrictions=1&lpa.peopleToNotify=1');
    });

    it("can submit the completed LPA", () => {
        cy.contains('h1', "Check your LPA")

        cy.checkA11yApp();

        cy.contains('h2', "LPA decisions")

        cy.contains('dt', "When your attorneys can use your LPA")
        cy.contains('dt', "Who is your attorney")
        cy.contains('dt', "Who are your replacement attorneys")

        cy.contains('h2', "People named on the LPA")
        cy.contains('h3', "Donor")
        cy.contains('h3', "Certificate provider")
        cy.contains('h3', "Attorneys")

        cy.get('#f-checked-and-happy').check()

        cy.contains('button', 'Confirm').click();

        cy.url().should('contain', '/lpa-details-saved');
    });

    it("errors when not selected", () => {
        cy.contains('button', 'Confirm').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select the box if you have checked your LPA and are happy to share it with your certificate provider');
        });

        cy.contains('.govuk-form-group .govuk-error-message', 'Select the box if you have checked your LPA and are happy to share it with your certificate provider');
    })
});
