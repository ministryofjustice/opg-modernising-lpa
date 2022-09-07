describe('Choose attorneys address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-attorneys-address');
    });

    it('address can be looked up', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-select-address').select('123 Fake Street, Someville, NG1');
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/want-replacement-attorneys');
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
        cy.url().should('contain', '/want-replacment-attorneys');
    });
});
