describe('Choose attorneys address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-attorneys-address?id=without-address&withIncompleteAttorneys=1');
    });
    
    it('address can be looked up', () => {
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
        cy.url().should('contain', '/choose-attorneys-summary');
    });

    it('address can be entered manually', () => {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('a', "Can not find address?").click();

        cy.get('#f-address-line-1').type('Flat 2');
        cy.get('#f-address-line-2').type('123 Fake Street');
        cy.get('#f-address-line-3').type('Pretendingham');
        cy.get('#f-address-town').type('Someville');
        cy.get('#f-address-postcode').type('NG1');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/choose-attorneys-summary');
    });
    
    it('errors when empty postcode', () => {
        cy.contains('button', 'Find address').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter postcode');
        });
        
        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter postcode');
    });

    it('errors when unselected', () => {        
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('button', 'Continue').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select address');
        });
        
        cy.contains('[for=f-select-address] + .govuk-error-message', 'Select address');
    });

    it('errors when manual incorrect', () => {        
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();
        cy.contains('a', "Can not find address?").click();
        cy.contains('button', 'Continue').click();
        
        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter address line 1');
            cy.contains('Enter town or city');
        });
        
        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Enter address line 1');
        cy.contains('[for=f-address-town] + .govuk-error-message', 'Enter town or city');

        cy.get('#f-address-line-1').invoke('val', 'a'.repeat(51));
        cy.get('#f-address-line-2').invoke('val', 'b'.repeat(51));
        cy.get('#f-address-line-3').invoke('val', 'c'.repeat(51));
        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-address-line-1] + .govuk-error-message', 'Address line 1 must be 50 characters or less');
        cy.contains('[for=f-address-line-2] + .govuk-error-message', 'Address line 2 must be 50 characters or less');
        cy.contains('[for=f-address-line-3] + .govuk-error-message', 'Address line 3 must be 50 characters or less');
    });
});
