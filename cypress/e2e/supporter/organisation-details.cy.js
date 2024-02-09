const { TestEmail } = require("../../support/e2e");

describe('Organisation details', () => {
    beforeEach(() => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/organisation-details&inviteMembers=1');
    });

    it('shows invited members', () => {
        cy.checkA11yApp();
        cy.get("#tab_team-members").click()

        cy.contains("Invited team members")

        cy.contains("dd", "kamalsingh@example.org").parent().within(() => {
            cy.contains("Kamal Singh")
            cy.contains("Invite pending")
            cy.contains("a", "Resend invite")
        })

        cy.contains("dd", "jo_alessi@example.org").parent().within(() => {
            cy.contains("Jo Alessi")
            cy.contains("Invite pending")
            cy.contains("a", "Resend invite")
        })
    });
});
