const { TestEmail } = require("../../support/e2e");

describe('Invite member', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/invite-member');
    });

    it('can invite a member', () => {
        cy.checkA11yApp();

        cy.get('#f-email').type(TestEmail);
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');

        cy.contains('button', 'Send invite').click();

        cy.url().should('contain', '/manage-organisation/manage-team-members');
        cy.checkA11yApp();

        cy.contains('.govuk-notification-banner--success', TestEmail);
    });

    it('can invite an admin', () => {
        cy.get('#f-email').type(TestEmail);
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('[name="permission"]').check('admin', { force: true });

        cy.contains('button', 'Send invite').click();

        cy.url().should('contain', '/manage-organisation/manage-team-members');
        cy.checkA11yApp();

        cy.contains('.govuk-notification-banner--success', TestEmail);
    });

    it('errors when empty', () => {
        cy.contains('button', 'Send invite').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter email address');
        });

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('[for=f-email] ~ .govuk-error-message', 'Enter email address');
    });

    it('errors when names too long', () => {
        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));

        cy.contains('button', 'Send invite').click();

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
    });

    it('errors when invalid email', () => {
        cy.get('#f-email').type('not-an-email');

        cy.contains('button', 'Send invite').click();

        cy.contains('[for=f-email] ~ .govuk-error-message', 'Email address must be in the correct format, like name@example.com');
    });
});
