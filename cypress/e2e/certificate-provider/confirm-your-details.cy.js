describe('Confirm your details', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/enter-date-of-birth&lpa.certificateProvider=1&certificateProviderProvided=1&loginAs=certificate-provider');

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.contains('button', 'Continue').click();
    });

    it('shows details', () => {
        cy.url().should('contain', '/confirm-your-details');
        cy.checkA11yApp();

        cy.contains('1 February 1990');
        cy.contains('Jessie Jones');
        cy.contains('5 RICHMOND PLACE');
        cy.contains('07700900000');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your details').should('contain', 'Completed');
    });
});
