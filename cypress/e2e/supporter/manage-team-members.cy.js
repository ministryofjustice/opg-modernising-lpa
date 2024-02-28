describe('Organisation details', () => {
    it('shows invited and joined members', () => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&invitedMembers=2&members=2&permission=admin');

        cy.checkA11yApp();
        cy.contains("a", "Manage team members").click()

        cy.contains("Invited team members")

        cy.contains("td", "kamal-singh@example.org").parent().within(() => {
            cy.contains("Kamal Singh")
            cy.contains("Invite pending")
        })

        cy.contains("td", "jo-alessi@example.org").parent().within(() => {
            cy.contains("Jo Alessi")
            cy.contains("Invite pending")
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

    it('shows resend invite when expired', () => {
        cy.visit('/fixtures/supporter?organisation=1&redirect=/manage-organisation/manage-team-members&invitedMembers=1&permission=admin&expireInvites=1');

        cy.checkA11yApp();
        cy.contains("a", "Manage team members").click()

        cy.contains("Invited team members")

        cy.contains("td", "kamal-singh@example.org").parent().within(() => {
            cy.contains("Invite expired")
            cy.contains("button", "Resend invite").click({force: true})
        })

        cy.url().should("contain", "/manage-organisation/manage-team-members");
        cy.checkA11yApp();

        cy.contains(".govuk-notification-banner--success", "kamal-singh@example.org");
        cy.get("main").should("not.contain", "Resend invite");
    });
});
