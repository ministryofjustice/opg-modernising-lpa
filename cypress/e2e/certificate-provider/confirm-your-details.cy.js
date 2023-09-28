describe('Confirm your details', () => {
    it('shows details', () => {
        cy.visit('/fixtures/certificate-provider?redirect=/enter-date-of-birth');

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-your-details');
        cy.checkA11yApp();

        cy.contains('1 February 1990');
        cy.contains('Charlie Cooper');
        cy.contains('5 RICHMOND PLACE');
        cy.contains('07700900000');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-role');
    });

    it('redirects to tasklist when LPA has already been witnessed', () => {
        cy.visit('/fixtures/certificate-provider?redirect=/confirm-your-details&progress=signedByDonor');

        cy.url().should('contain', '/confirm-your-details');
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your details').should('contain', 'Completed');
    });

    it('redirects to tasklist when details have already been confirmed', () => {
        cy.visit('/fixtures/certificate-provider?redirect=/confirm-your-details&progress=detailsConfirmed');

        cy.url().should('contain', '/confirm-your-details');
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your details').should('contain', 'Completed');
    });
});
