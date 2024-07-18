describe('Edit member', () => {
    describe('admin', () => {
        beforeEach(() => {
            cy.visit("/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&members=2&asMember=alice-moxom@example.org&permission=admin");
            cy.url().should('contain', "/manage-organisation/manage-team-members");
        })

        it('can edit a team members name', () => {
            cy.contains('a', "Leon Vynehall").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();

            cy.get('#f-first-names').clear().type('John');
            cy.get('#f-last-name').clear().type('Doe');

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members");

            cy.checkA11yApp();

            cy.contains('Team member’s name updated to John Doe');
            cy.contains('a', "John Doe")
        })

        it('can edit own name', () => {
            cy.contains('a', "Alice Moxom").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();

            cy.get('#f-first-names').clear().type('John');
            cy.get('#f-last-name').clear().type('Doe');

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members");

            cy.checkA11yApp();

            cy.contains('Your name has been updated to John Doe');
            cy.contains('a', "John Doe")
        })

        it('can update a team members access to the organisation', () => {
            cy.contains('a', "Leon Vynehall").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();

            cy.get('[name="status"]').check('suspended', { force: true });

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members");

            cy.checkA11yApp();

            cy.contains('leon-vynehall@example.org has been suspended from this organisation.');
            cy.contains("td", "leon-vynehall@example.org").parent().contains("Suspended")

            cy.contains('a', "Leon Vynehall").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();

            cy.get('[name="status"]').check('active', { force: true });

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members");

            cy.checkA11yApp();

            cy.contains('leon-vynehall@example.org can now access this organisation.');
            cy.contains("td", "leon-vynehall@example.org").parent().contains("Active")
        })

        it('can not update own access to the organisation', () => {
            cy.contains('a', "Alice Moxom").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();

            cy.get('[name="status"]').should('not.exist');
        })

        it('multiple update banners are stacked', () => {
            cy.visit("/supporter/manage-organisation/manage-team-members?statusUpdated=suspended&statusEmail=a@b.com&nameUpdated=A+B");

            cy.checkA11yApp();

            cy.contains('Team member’s name updated to A B.');
            cy.contains('a@b.com has been suspended from this organisation.');
        })
    })

    describe('non-admin', () => {
        it('can edit own name', () => {
            cy.visit("/fixtures/supporter?organisation=1&redirect=/dashboard&members=1&asMember=alice-moxom@example.org");

            cy.contains('a', 'Manage your details').click();
            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");

            cy.checkA11yApp();
            cy.contains('Your name');

            cy.get('#f-first-names').clear().type('John');
            cy.get('#f-last-name').clear().type('Doe');

            cy.contains('button', "Save").click()

            cy.url().should('contain', "/dashboard");

            cy.checkA11yApp();
            cy.contains('Your name has been updated to John Doe');
        })
    })

    describe('errors', () => {
        beforeEach(() => {
            cy.visit("/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&members=1");

            cy.url().should('contain', "/manage-organisation/manage-team-members");
            cy.contains('a', "Alice Moxom").click()

            cy.url().should('contain', "/manage-organisation/manage-team-members/edit-team-member");
        });

        it('errors when empty', () => {
            cy.get('#f-first-names').clear();
            cy.get('#f-last-name').clear();
            cy.get('#f-status').invoke('attr', 'checked', false);

            cy.contains('button', "Save").click()

            cy.checkA11yApp();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter first names');
                cy.contains('Enter last name');
                cy.contains('Select status');
            });

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
            cy.contains('#status-error', 'Select status');
        });

        it('errors when names too long', () => {
            cy.get('#f-first-names').invoke('val', 'a '.repeat(54));
            cy.get('#f-last-name').invoke('val', 'b '.repeat(62));

            cy.contains('button', "Save").click()

            cy.checkA11yApp();

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
        });
    });
})
