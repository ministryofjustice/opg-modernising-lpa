describe('Read the LPA', () => {
    describe('when the LPA is signed', () => {
        beforeEach(() => {
            cy.visit('/fixtures/certificate-provider?redirect=/read-the-lpa&progress=confirmYourIdentity');
        });

        it('displays the LPA details and goes to provide certificate', () => {
            cy.checkA11yApp();

            cy.contains('Donor: Sam Smith');
            cy.contains('Certificate provider: Charlie Cooper');
            cy.contains('Attorney: Jessie Jones');
            cy.contains('Attorney: Robin Redcar');
            cy.contains('Signed by Sam Smith on: 2 January 2023');
            cy.contains('Witnessed by Charlie Cooper on: 2 January 2023');

            cy.contains('button', 'Continue').click();
            cy.url().should('contain', '/what-happens-next');
        });
    });

    describe('when the LPA is not yet signed', () => {
        beforeEach(() => {
            cy.visit('/fixtures/certificate-provider?redirect=/read-the-draft-lpa');
        });

        it('displays the LPA details and goes to task list', () => {
            cy.checkA11yApp();

            cy.contains('Donor: Sam Smith');
            cy.contains('Certificate provider: Charlie Cooper');
            cy.contains('Attorney: Jessie Jones');
            cy.contains('Attorney: Robin Redcar');

            cy.contains('button', 'Continue').should('not.exist');
            cy.contains('a', 'Return to task list').click();
            cy.url().should('contain', '/task-list');
        });
    });

    describe('when on task list in other language', () => {
        beforeEach(() => {
            cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=confirmYourIdentity');
            cy.contains('a', 'Cymraeg').click();
        });

        it('displays the LPA details in the registration language first', () => {
            cy.contains('a', 'Darparu eich tystysgrif').click();
            cy.contains('h1', 'Read the LPA');
        });
    });
});
