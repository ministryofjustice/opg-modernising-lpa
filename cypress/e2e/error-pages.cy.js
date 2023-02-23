describe('Error pages', () => {
    it('shows for 404s', () => {
        cy.visit('/not-a-real-page', {failOnStatusCode: false});
        cy.contains('Page not found');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.visit('/lpa', {failOnStatusCode: false});
        cy.contains('Page not found');

        cy.visit('/lpa/1000', {failOnStatusCode: false});
        cy.contains('Page not found');

        cy.visit('/testing-start?redirect=/not-a-real-page', {failOnStatusCode: false});
        cy.contains('Page not found');
    });

    it('shows for 500s', () => {
        cy.visit('/testing-start?redirect=/task-list');
        cy.visitLpa('/payment-confirmation', { failOnStatusCode: false });

        cy.contains('Sorry, there is a problem with the service');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });
    });

    it('shows for invalid CSRF tokens', () => {
        cy.visit('/testing-start?redirect=/your-details');
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
