const { TestEmail } = require("../../support/e2e");

describe('Organisation details', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&inviteMembers=1');
    });

    it('shows invited members', () => {
        cy.checkA11yApp();
        cy.contains("a", "Manage team members").click()

        cy.contains("Invited team members")

        cy.contains("td", "kamalsingh@example.org").parent().within(() => {
            cy.contains("Kamal Singh")
            cy.contains("Invite pending")
            cy.contains("a", "Resend invite")
        })

        cy.contains("td", "jo_alessi@example.org").parent().within(() => {
            cy.contains("Jo Alessi")
            cy.contains("Invite pending")
            cy.contains("a", "Resend invite")
        })
    });
});
