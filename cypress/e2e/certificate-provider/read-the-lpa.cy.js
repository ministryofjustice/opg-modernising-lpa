describe('Read the LPA', () => {
    describe('when the LPA is signed', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/read-the-lpa&lpa.complete=1&certificateProviderProvided=1&loginAs=certificate-provider');
        });

        it('displays the LPA details and goes to provide certificate', () => {
            cy.checkA11yApp();

            cy.contains('dt', "When attorneys can use the LPA")
            cy.contains('dt', "Their attorneys")
            cy.contains('dt', "Their replacement attorneys")

            cy.contains('Continue').click();
            cy.url().should('contain', '/what-happens-next');
        });
    });

    describe('when the LPA is not yet signed', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/read-the-lpa&lpa.certificateProvider=1&lpa.yourDetails=1&certificateProviderProvided=1&loginAs=certificate-provider');
        });

        it('displays the LPA details and goes to task list', () => {
            cy.checkA11yApp();

            cy.contains('dt', "When attorneys can use the LPA")
            cy.contains('dt', "Their attorneys")
            cy.contains('dt', "Their replacement attorneys")

            cy.contains('Continue').click();
            cy.url().should('contain', '/task-list');
        });
    });
});
