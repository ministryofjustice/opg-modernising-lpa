describe('People to notify address', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-people-to-notify');

        cy.get('#f-first-names').type('a');
        cy.get('#f-last-name').type('b');
        cy.get('#f-email').type('a.b@example.com');
        cy.contains('button', 'Continue').click();
    });

    it('errors when empty postcode', () => {
        cy.contains('button', 'Find address').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter a postcode');
        });

        cy.contains('[for=f-lookup-postcode] + .govuk-error-message', 'Enter a postcode');
    });

    it('errors when unselected', () => {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Select an address from the list');
        });

        cy.contains('[for=f-select-address] + .govuk-error-message', 'Select an address from the list');
    });

    it('errors when manual incorrect', () => {
        cy.get('#f-lookup-postcode').type('NG1');
        cy.contains('button', 'Find address').click();
        cy.contains('a', "I can’t find their address in the list").click();
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
