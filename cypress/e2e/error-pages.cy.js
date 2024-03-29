describe('Error pages', () => {
    it('shows for 404s', () => {
        cy.visit('/not-a-real-page', { failOnStatusCode: false });
        cy.contains('Page not found');
        cy.checkA11yApp();

        cy.visit('/lpa', { failOnStatusCode: false });
        cy.contains('Page not found');

        cy.visit('/lpa/1000', { failOnStatusCode: false });
        cy.contains('Page not found');

        cy.visit('/fixtures?redirect=/not-a-real-page', { failOnStatusCode: false });
        cy.contains('Page not found');
    });

    it('shows for 500s', () => {
        cy.visit('/fixtures?redirect=/task-list');
        cy.visitLpa('/payment-confirmation', { failOnStatusCode: false });

        cy.contains('Sorry, there is a problem with the service');
        cy.checkA11yApp();
    });

    it('shows for invalid CSRF tokens', () => {
        cy.visit('/fixtures?redirect=/your-details');
        cy.clearCookie('csrf');

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.contains('button', 'Continue').click();

        cy.contains('Sorry, there is a problem with the service');
    });
});
