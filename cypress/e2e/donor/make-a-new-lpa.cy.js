describe('Make a new LPA', () => {
    it('can create another reusing some previous details', () => {
        cy.visit('/fixtures/dashboard?asDonor=1&redirect=/dashboard');
        cy.contains('button', 'Start now').click();
        cy.checkA11yApp();

        // Name
        cy.contains("dd", "Sam Smith").parent().contains("a", "Change").click();

        cy.url().should('contain', '/your-name');
        cy.checkA11yApp();
        cy.get('#f-first-names').should('have.value', 'Sam').clear().type("a");
        cy.get('#f-last-name').should('have.value', 'Smith').clear().type("b");
        cy.contains("button", "Continue").click();

        cy.url().should('contain', '/we-have-updated-your-details');
        cy.checkA11yApp();
        cy.get(".govuk-notification-banner__header").contains("Name updated");
        cy.contains("a", "Continue").click();

        cy.url().should('contain', '/make-a-new-lpa');
        cy.contains("dd", "a b");

        // Date of birth
        cy.contains("dd", "2 January 2000").parent().contains("a", "Change").click()

        cy.url().should('contain', '/your-date-of-birth');
        cy.checkA11yApp();
        cy.get('#f-date-of-birth-year').should('have.value', '2000').clear().type("2001");
        cy.get('#f-date-of-birth-month').should('have.value', '1').clear().type("2");
        cy.get('#f-date-of-birth').should('have.value', '2').clear().type("3");
        cy.contains("button", "Continue").click();

        cy.url().should('contain', '/we-have-updated-your-details');
        cy.checkA11yApp();
        cy.get(".govuk-notification-banner__header").contains("Date of birth updated")
        cy.get('main').should('not.contain', 'contact OPG');
        cy.contains("a", "Continue").click();

        cy.url().should('contain', '/make-a-new-lpa');
        cy.contains("dd", "3 February 2001");

        // Address
        cy.contains("dd", "1 RICHMOND PLACE").parent().contains("a", "Change").click()

        cy.url().should('contain', '/your-address');
        cy.checkA11yApp();
        cy.get('#f-address-line-1').should('have.value', '1 RICHMOND PLACE').clear().type("2 RICHMOND PLACE")
        cy.get('#f-address-line-2').should('have.value', 'KINGS HEATH').clear();
        cy.get('#f-address-line-3').should('have.value', 'WEST MIDLANDS').clear()
        cy.get('#f-address-town').should('have.value', 'BIRMINGHAM');
        cy.get('#f-address-postcode').should('have.value', 'B14 7ED');
        cy.contains("button", "Continue").click();

        cy.url().should('contain', '/we-have-updated-your-details');
        cy.checkA11yApp();
        cy.get(".govuk-notification-banner__header").contains("Address updated");
        cy.contains("a", "Continue").click();

        cy.url().should('contain', '/make-a-new-lpa');
        cy.contains("dd", "2 RICHMOND PLACE");
        cy.get('main').should('not.contain', 'KINGS HEATH');
        cy.get('main').should('not.contain', 'WEST MIDLANDS');
        cy.contains("a", "Continue").click();

        cy.url().should('contain', '/can-you-sign-your-lpa');
    });
})
