describe('Organisation details', () => {
    it('shows invited and joined members', () => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&inviteMembers=2&members=2');

        cy.checkA11yApp();
        cy.contains("a", "Manage team members").click()

        cy.contains("Invited team members")

        cy.contains("td", "kamal-singh@example.org").parent().within(() => {
            cy.contains("Kamal Singh")
            cy.contains("Invite pending")
            cy.contains("a", "Resend invite")
        })

        cy.contains("td", "jo-alessi@example.org").parent().within(() => {
            cy.contains("Jo Alessi")
            cy.contains("Invite pending")
            cy.contains("a", "Resend invite")
        })

        cy.contains("Team members")

        cy.contains("td", "alice-moxom@example.org").parent().within(() => {
            cy.contains("Alice Moxom")
            cy.contains("Admin")
            cy.contains("Active")
        })

        cy.contains("td", "leon-vynehall@example.org").parent().within(() => {
            cy.contains("Leon Vynehall")
            cy.contains("Admin")
            cy.contains("Active")
        })
    });
});
