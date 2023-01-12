describe('Choose replacement attorneys address', () => {
    it('address can be looked up', () => {
        cy.visit('/testing-start?redirect=/choose-replacement-attorneys-address?id=without-address&withIncompleteAttorneys=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').should('have.value', '2 RICHMOND PLACE');
        cy.get('#f-address-line-2').should('have.value', '');
        cy.get('#f-address-line-3').should('have.value', '');
        cy.get('#f-address-town').should('have.value', 'BIRMINGHAM');
        cy.get('#f-address-postcode').should('have.value', 'B14 7ED');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-summary');
    });

    it('address can be entered manually', () => {
        cy.visit('/testing-start?redirect=/choose-replacement-attorneys-address?id=without-address&withIncompleteAttorneys=1');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('a', "I can’t find their address in the list").click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').type('Flat 2');
        cy.get('#f-address-line-2').type('123 Fake Street');
        cy.get('#f-address-line-3').type('Pretendingham');
        cy.get('#f-address-town').type('Someville');
        cy.get('#f-address-postcode').type('NG1');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-replacement-attorneys-summary');
    });
});
