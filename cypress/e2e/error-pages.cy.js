describe('Error pages', () => {
    it('shows for 404s', () => {
        cy.visit('/not-a-real-page', { failOnStatusCode: false });
        cy.contains('Page not found');
        cy.contains('Welsh').should('not.exist');
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
        cy.contains('Welsh').should('not.exist');
        cy.checkA11yApp();
    });

    it('shows for invalid CSRF tokens', () => {
        cy.visit('/fixtures?redirect=/your-name');
        cy.clearCookie('csrf');

        cy.get('#f-first-names').invoke('val', 'John');
        cy.get('#f-last-name').invoke('val', 'Doe');

        cy.contains('button', 'Save and continue').click();

        cy.contains('Sorry, there is a problem with the service');
        cy.contains('Welsh').should('not.exist');
    });
});
