describe('Edit member', () => {
    describe('admin', () => {
        beforeEach(() => {
            cy.visit("/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&members=1");

            cy.url().should('contain', "/manage-organisation/manage-team-members");
            cy.contains('a', "Alice Moxom").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");
        });

        it('can edit a team members name', () => {
            cy.checkA11yApp();

            cy.get('#f-first-names').clear().type('John');
            cy.get('#f-last-name').clear().type('Doe');

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members");

            cy.contains('Team member’s name updated to John Doe');
            cy.contains('a', "John Doe")
        })

        it('can edit own name', () => {
            // TODO update to a full test when admins can set their own names during org creation
            cy.visit("/supporter/manage-organisation/manage-team-members?nameUpdated=John+Doe&selfUpdated=1");

            cy.contains('Your name has been updated to John Doe');
        })

        it('errors when empty', () => {
            cy.get('#f-first-names').clear();
            cy.get('#f-last-name').clear();

            cy.contains('button', "Save").click()

            cy.checkA11yApp();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter first names');
                cy.contains('Enter last name');
            });

            cy.contains('[for=f-first-names] + .govuk-error-message', 'Enter first names');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        });

        it('errors when names too long', () => {
            cy.get('#f-first-names').invoke('val', 'a '.repeat(54));
            cy.get('#f-last-name').invoke('val', 'b '.repeat(62));

            cy.contains('button', "Save").click()

            cy.checkA11yApp();

            cy.contains('[for=f-first-names] + .govuk-error-message', 'First names must be 53 characters or less');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
        });
    })

    describe('non-admin', () => {
        it.only('can edit own name', () => {
            cy.visit("/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&members=1&asMember=alice-moxom@example.org");

            cy.contains('a', 'Manage your details').click();
            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();
            cy.contains('Your name');

            cy.get('#f-first-names').clear ().type('John');
            cy.get('#f-last-name').clear().type('Doe');

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/dashboard");

            cy.contains('Your name has been updated to John Doe');
        })
    })

})
