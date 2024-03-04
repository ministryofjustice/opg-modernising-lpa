describe('Manage organisation', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/organisation-details&lpa=5');
        cy.checkA11yApp();
    });

    it('name can be changed', () => {
        cy.contains('a', 'Change').click()

        cy.url().should('contain', '/manage-organisation/organisation-details/edit-organisation-name');
        cy.checkA11yApp();

        cy.get('#f-name').clear().type('My organisation');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/manage-organisation/organisation-details');
        cy.checkA11yApp();
        cy.contains('Your organisation name has been saved.');
        cy.contains('.govuk-summary-list', 'My organisation');
    });

    it('can be deleted', () => {
        cy.get('span.app-service-header-caption')
            .invoke('text')
            .then((text) => {
                const orgName = text.trim();

                cy.contains('a', 'Delete organisation').click()

                cy.url().should('contain', '/manage-organisation/organisation-details/delete-organisation');
                cy.checkA11yApp();

                cy.contains('5 LPAs that are still in progress')

                cy.contains('button', 'Delete organisation').click()

                cy.url().should('contain', '/organisation-deleted');
                cy.checkA11yApp();

                cy.contains(`The organisation ${orgName} has been deleted`)
            });
    })
});
