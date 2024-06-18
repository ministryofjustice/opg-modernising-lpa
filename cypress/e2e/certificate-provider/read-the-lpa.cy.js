describe('Read the LPA', () => {
    describe('when the LPA is signed', () => {
        beforeEach(() => {
            cy.visit('/fixtures/certificate-provider?redirect=/read-the-lpa&progress=confirmYourIdentity');
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
            cy.visit('/fixtures/certificate-provider?redirect=/read-the-lpa');
        });

        it('displays the LPA details and goes to task list', () => {
            cy.checkA11yApp();

            cy.contains('dt', "When attorneys can use the LPA")
            cy.contains('dt', "Their attorneys")
            cy.contains('dt', "Their replacement attorneys")

            cy.get('button').should('not.contain', 'Continue');
            cy.contains('Return to task list').click();
            cy.url().should('contain', '/task-list');
        });
    });
});
