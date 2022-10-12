describe('Donor address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/your-address');
    });

    it('address can be looked up', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B147ED');
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/who-is-the-lpa-for');
    });

    it('address can be entered manually', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('a', "Can't find address?").click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').type('Flat 2');
        cy.get('#f-address-line-2').type('123 Fake Street');
        cy.get('#f-address-town').type('Someville');
        cy.get('#f-address-postcode').type('NG1');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/who-is-the-lpa-for');
    });
});
