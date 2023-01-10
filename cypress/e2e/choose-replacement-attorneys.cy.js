describe('Choose replacement attorneys', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-replacement-attorneys');
        cy.injectAxe();
    });

    it('can be submitted', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type('name@example.com');
        cy.get('#f-date-of-birth-day').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-address');
        cy.contains("John Doe's address");
    });
});
